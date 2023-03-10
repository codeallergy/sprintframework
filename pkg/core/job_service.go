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

package core

import (
	"context"
	"github.com/pkg/errors"
	"github.com/codeallergy/sprint"
	"go.uber.org/zap"
	"strings"
	"sync"
)

var ErrJobNotFound = errors.New("job not found")

type implJobService struct {
	Log           *zap.Logger              `inject`

	muJobs  sync.Mutex
	jobs    []*sprint.JobInfo
}

func JobService() sprint.JobService {
	return &implJobService{}
}

func (t *implJobService) ListJobs() ([]string, error) {
	t.muJobs.Lock()
	defer t.muJobs.Unlock()

	var list []string
	for _, job := range t.jobs {
		list = append(list, job.Name)
	}

	return list, nil
}

func (t *implJobService) AddJob(job *sprint.JobInfo) error {
	t.muJobs.Lock()
	defer t.muJobs.Unlock()

	t.jobs = append(t.jobs, job)
	return nil
}

func (t *implJobService) CancelJob(name string) error {
	t.muJobs.Lock()
	defer t.muJobs.Unlock()

	for i, job := range t.jobs {
		if job.Name == name {
			t.jobs = append(t.jobs[:i], t.jobs[i+1:]...)
			return nil
		}
	}

	return ErrJobNotFound
}

func (t *implJobService) RunJob(ctx context.Context, name string) (err error) {

	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				err = v
			case string:
				err = errors.New(v)
			default:
				err = errors.Errorf("%v", v)
			}
		}
	}()

	job, err := t.findJob(name)
	if err != nil {
		return err
	}

	return job.ExecutionFn(ctx)
}

func (t *implJobService) findJob(name string) (*sprint.JobInfo, error) {
	t.muJobs.Lock()
	defer t.muJobs.Unlock()

	for _, job := range t.jobs {
		if job.Name == name {
			return job, nil
		}
	}

	return nil, ErrJobNotFound
}


func (t *implJobService) ExecuteCommand(cmd string, args []string) (string, error) {

	switch cmd {
	case "list":
		list, err := t.ListJobs()
		if err != nil {
			return "", err
		}
		return strings.Join(list, "\n"), nil

	case "run":
		if len(args) < 1 {
			return "Usage: job run name", nil
		}
		jobName := args[0]
		go func() {

			err := t.RunJob(context.Background(), jobName)
			if err != nil {
				t.Log.Error("JobRun", zap.String("jobName", jobName), zap.Error(err))
			}

		}()
		return "OK", nil

	case "cancel":
		if len(args) < 1 {
			return "Usage: job cancel name", nil
		}
		jobName := args[0]
		err := t.CancelJob(jobName)
		if err != nil {
			return "", errors.Errorf("cancel of job '%s' was failed, %v", jobName, err)
		}
		return"OK", nil

	default:
		return "", errors.Errorf("unknown job command '%s'", cmd)
	}

}