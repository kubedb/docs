/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package user

import (
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"

	"golang.org/x/crypto/bcrypt"
)

// returns true if the user exists.
// otherwise false.
func HasUser(userList map[string]api.ElasticsearchUserSpec, username api.ElasticsearchInternalUser) bool {
	if _, exist := userList[string(username)]; exist {
		return true
	}
	return false
}

// Set user if missing
func SetMissingUser(userList map[string]api.ElasticsearchUserSpec, username api.ElasticsearchInternalUser, userSpec api.ElasticsearchUserSpec) {
	if HasUser(userList, username) {
		return
	}

	userList[string(username)] = userSpec
}

func SetPasswordHashForUser(userList map[string]api.ElasticsearchUserSpec, username string, password string) error {
	var userSpec api.ElasticsearchUserSpec
	if value, exist := userList[username]; exist {
		userSpec = value
	}

	hash, err := generatePasswordHash(password)
	if err != nil {
		return err
	}

	userSpec.Hash = hash
	userList[username] = userSpec
	return nil
}

func generatePasswordHash(password string) (string, error) {
	pHash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(pHash), nil
}
