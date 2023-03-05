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

package app

import (
	"github.com/codeallergy/glue"
	"github.com/codeallergy/sprintframework/pkg/assets"
	"github.com/codeallergy/sprintframework/pkg/assetsgz"
	"github.com/codeallergy/sprintframework/pkg/resources"
	"os"
)

var DefaultFileModes = map[string]interface{} {
	"log.dir": os.FileMode(0775),
	"log.file": os.FileMode(0664),
	"backup.file": os.FileMode(0664),
	"exe.file": os.FileMode(0775),
	"run.dir": os.FileMode(0775),
	"pid.file": os.FileMode(0666),
	"data.dir": os.FileMode(0770),
	"data.file": os.FileMode(0664),
}

var DefaultResources = &glue.ResourceSource{
	Name: "resources",
	AssetNames: resources.AssetNames(),
	AssetFiles: resources.AssetFile(),
}

var DefaultAssets = &glue.ResourceSource{
	Name: "assets",
	AssetNames: assets.AssetNames(),
	AssetFiles: assets.AssetFile(),
}

var DefaultGzipAssets = &glue.ResourceSource{
	Name: "assets-gzip",
	AssetNames: assetsgz.AssetNames(),
	AssetFiles: assetsgz.AssetFile(),
}

var DefaultApplicationBeans = []interface{}{
	ApplicationFlags(100000), // override any property resolvers
	FlagSetFactory(),
	ResourceService(),
}
