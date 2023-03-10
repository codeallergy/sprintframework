/*
 * Copyright (c) 2022-2023 Zander Schwid & Co. LLC.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 */

package server

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/codeallergy/glue"
	"github.com/codeallergy/sprint"
	"github.com/codeallergy/sprintframework/pkg/util"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
	"os/user"
	"strings"
	"sync"
	"time"
)

type userMetadataKey struct{}

type implAuthorizationMiddleware struct {
	Application      sprint.Application      `inject`
	Properties       glue.Properties         `inject`
	ConfigRepository sprint.ConfigRepository `inject`
	Log              *zap.Logger              `inject`

	invalidTokens     sync.Map   // key is string, value is true

	secretKey []byte   // JWT tokens secret key
}

func AuthorizationMiddleware() sprint.AuthorizationMiddleware {
	return &implAuthorizationMiddleware{}
}

func (t *implAuthorizationMiddleware) PostConstruct() (err error) {

	secret := t.Properties.GetString("jwt.secret.key", "")

	if secret == "" {
		secret, err = util.GenerateToken()
		if err != nil {
			return err
		}

		fmt.Printf("Generated JWT 'jwt.secret.key' property: %s\n", secret)
		err = t.ConfigRepository.Set("jwt.secret.key", secret)
		if err != nil {
			return err
		}

		authToken, err := t.generateDefaultAuthToken(secret)
		if err != nil {
			return err
		}
		fmt.Printf("export %s_AUTH=%s\n", strings.ToUpper(t.Application.Name()), authToken)
	}

	t.secretKey, err = base64.RawURLEncoding.DecodeString(secret)
	return err
}

func (t *implAuthorizationMiddleware) generateDefaultAuthToken(secret string) (string, error) {

	secretKey, err := base64.RawURLEncoding.DecodeString(secret)
	if err != nil {
		return "", err
	}

	user, err := user.Current()
	if err != nil {
		return "", err
	}

	u := &sprint.AuthorizedUser{
		Username:  user.Username,
		Roles:     map[string]bool {
			"USER": true,
			"ADMIN": true,
		},
		Context:   make(map[string]string),
		ExpiresAt: time.Now().Unix() + 356*24*3600,
	}

	return util.GenerateAuthToken(secretKey, u)
}

func (t *implAuthorizationMiddleware) Authenticate(ctx context.Context) (outCtx context.Context, err error) {

	user, ok := t.doAuthenticate(ctx)
	if !ok {

		user = &sprint.AuthorizedUser{
			Username:  "",
			Roles:     nil,
			Context:   nil,
			ExpiresAt: 0,
			Token:     "",
		}
	}

	return context.WithValue(ctx, userMetadataKey{}, user), nil

}

func (t *implAuthorizationMiddleware) doAuthenticate(ctx context.Context) (*sprint.AuthorizedUser, bool) {

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, false
	}

	authHeaders, ok := md["authorization"]
	if !ok {
		return nil, false
	}

	if len(authHeaders) != 1 {
		return nil, false
	}

	auth := authHeaders[0]

	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return nil, false
	}

	token := strings.TrimSpace(strings.TrimPrefix(auth, prefix))
	if token == "" {
		return nil, false
	}

	if _, ok := t.invalidTokens.Load(token); ok {
		return nil, false
	}

	user, err := util.VerifyAuthToken(t.secretKey, token)
	if err != nil {
		return nil, false
	}

	return user, true
}

func (t *implAuthorizationMiddleware) GetUser(ctx context.Context) (*sprint.AuthorizedUser, bool) {
	userMetadata := ctx.Value(userMetadataKey{})
	if user, ok := userMetadata.( *sprint.AuthorizedUser); ok {
		if user.Username == "" && user.ExpiresAt == 0 {
			return nil, false
		} else {
			return user, true
		}
	} else {
		// Middleware could miss the request
		return t.doAuthenticate(ctx)
	}
}

func (t *implAuthorizationMiddleware) HasUserRole(ctx context.Context, role string) bool {
	user, ok := t.GetUser(ctx)
	if !ok || user.Roles == nil {
		return false
	}
	return user.Roles[role]
}

func (t *implAuthorizationMiddleware) UserContext(ctx context.Context, name string) (string, bool) {
	user, ok := t.GetUser(ctx)
	if !ok || user.Context == nil {
		return "", false
	}
	value, ok := user.Context[name]
	return value, ok
}

func (t *implAuthorizationMiddleware) GenerateToken(user *sprint.AuthorizedUser) (string, error) {
	return util.GenerateAuthToken(t.secretKey, user)
}

func (t *implAuthorizationMiddleware) ParseToken(token string) (*sprint.AuthorizedUser, error) {
	return util.VerifyAuthToken(t.secretKey, token)
}

func (t *implAuthorizationMiddleware) InvalidateToken(token string) {
	t.invalidTokens.Store(token, true)
}

