/*
* Flow CLI
*
* Copyright 2019-2021 Dapper Labs, Inc.
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*   http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */
package config_test

import (
	"os"
	"testing"

	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/config/json"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

var mockFS = afero.NewMemMapFs()
var af = afero.Afero{Fs: mockFS}

func Test_JSONSimple(t *testing.T) {
	b := []byte(`{
		"emulators": {
			"default": {
				"port": 3569,
				"serviceAccount": "emulator-account"
			}
		},
		"contracts": {},
		"networks": {
			"emulator": "127.0.0.1:3569"
		},
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "21c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
			}
		},
		"deployments": {}
	}`)

	err := afero.WriteFile(mockFS, "test2-flow.json", b, 0644)

	assert.NoError(t, err)

	composer := config.NewLoader(af)
	composer.AddConfigParser(json.NewParser())
	conf, loadErr := composer.Load([]string{"test2-flow.json"})

	assert.NoError(t, loadErr)
	assert.Equal(t, 1, len(conf.Accounts))
	assert.Equal(t, "0x21c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7", conf.Accounts[0].Key.PrivateKey.String())
}

func Test_ErrorWhenMissingBothDefaultJsonFiles(t *testing.T) {
	composer := config.NewLoader(afero.Afero{Fs: mockFS})
	composer.AddConfigParser(json.NewParser())

	_, loadErr := composer.Load(config.DefaultPaths())

	assert.Error(t, loadErr)
	assert.Contains(t, loadErr.Error(), "missing configuration")
}

func Test_AllowMissingocalJson(t *testing.T) {
	b := []byte(`{
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "21c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
			}
		}
	}`)

	mockFS := afero.NewMemMapFs()
	err := afero.WriteFile(mockFS, config.GlobalPath(), b, 0644)

	assert.NoError(t, err)

	composer := config.NewLoader(afero.Afero{Fs: mockFS})
	composer.AddConfigParser(json.NewParser())

	conf, loadErr := composer.Load(config.DefaultPaths())

	assert.NoError(t, loadErr)
	assert.Equal(t, 1, len(conf.Accounts))
	assert.Equal(t, "0x21c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7",
		conf.Accounts.ByName("emulator-account").Key.PrivateKey.String(),
	)
}

func Test_PreferLocalJson(t *testing.T) {
	b := []byte(`{
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "21c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
			}
		}
	}`)

	b2 := []byte(`{
		 "accounts":{
				"emulator-account":{
					 "address":"f1d6e0586b0a20c7",
					 "key":"3335dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
				}
		 }
	}`)

	mockFS := afero.NewMemMapFs()
	err := afero.WriteFile(mockFS, "flow.json", b, 0644)
	err2 := afero.WriteFile(mockFS, config.GlobalPath(), b2, 0644)

	assert.NoError(t, err)
	assert.NoError(t, err2)

	composer := config.NewLoader(afero.Afero{Fs: mockFS})
	composer.AddConfigParser(json.NewParser())

	conf, loadErr := composer.Load(config.DefaultPaths())

	assert.NoError(t, loadErr)
	assert.Equal(t, 1, len(conf.Accounts))
	assert.Equal(t, "0x21c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7",
		conf.Accounts.ByName("emulator-account").Key.PrivateKey.String(),
	)
}

func Test_ComposeJSON(t *testing.T) {
	b := []byte(`{
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "21c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
			}
		}
	}`)

	b2 := []byte(`{
		 "accounts":{
				"admin-account":{
					 "address":"f1d6e0586b0a20c7",
					 "key":"3335dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
				}
		 }
	}`)

	mockFS := afero.NewMemMapFs()
	err := afero.WriteFile(mockFS, "flow.json", b, 0644)
	err2 := afero.WriteFile(mockFS, "flow-testnet.json", b2, 0644)

	assert.NoError(t, err)
	assert.NoError(t, err2)

	composer := config.NewLoader(afero.Afero{Fs: mockFS})
	composer.AddConfigParser(json.NewParser())

	conf, loadErr := composer.Load([]string{"flow.json", "flow-testnet.json"})

	assert.NoError(t, loadErr)
	assert.Equal(t, 2, len(conf.Accounts))
	assert.Equal(t, "0x21c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7",
		conf.Accounts.ByName("emulator-account").Key.PrivateKey.String(),
	)
	assert.NotNil(t, conf.Accounts.ByName("admin-account"))
	assert.Equal(t, "0x3335dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7",
		conf.Accounts.ByName("admin-account").Key.PrivateKey.String(),
	)
}

func Test_ComposeJSONOverwrite(t *testing.T) {
	b := []byte(`{
		"accounts": {
			"admin-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "21c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
			}
		}
	}`)

	b2 := []byte(`{
		 "accounts":{
				"admin-account":{
					 "address":"f1d6e0586b0a20c7",
					 "key":"3335dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
				}
		 }
	}`)

	mockFS := afero.NewMemMapFs()
	err := afero.WriteFile(mockFS, "flow.json", b, 0644)
	err2 := afero.WriteFile(mockFS, "flow-testnet.json", b2, 0644)

	assert.NoError(t, err)
	assert.NoError(t, err2)

	composer := config.NewLoader(afero.Afero{Fs: mockFS})
	composer.AddConfigParser(json.NewParser())

	conf, loadErr := composer.Load([]string{"flow.json", "flow-testnet.json"})

	assert.NoError(t, loadErr)
	assert.Equal(t, 1, len(conf.Accounts))
	assert.NotNil(t, conf.Accounts.ByName("admin-account"))
	assert.Equal(t, "0x3335dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7",
		conf.Accounts.ByName("admin-account").Key.PrivateKey.String(),
	)
}

func Test_FromFileAccountSimple(t *testing.T) {
	b := []byte(`{
		"accounts": {
			"service-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "21c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
			},
			"admin-account": { "fromFile": "private.json" }
		}
	}`)

	b2 := []byte(`{
		 "accounts":{
				"admin-account":{
					 "address":"f1d6e0586b0a20c7",
					 "key":"3335dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
				}
		 }
	}`)

	mockFS := afero.NewMemMapFs()
	err := afero.WriteFile(mockFS, "flow.json", b, 0644)
	err2 := afero.WriteFile(mockFS, "private.json", b2, 0644)

	assert.NoError(t, err)
	assert.NoError(t, err2)

	composer := config.NewLoader(afero.Afero{Fs: mockFS})
	composer.AddConfigParser(json.NewParser())

	conf, loadErr := composer.Load([]string{"flow.json", "private.json"})

	assert.NoError(t, loadErr)
	assert.Equal(t, 2, len(conf.Accounts))
	assert.NotNil(t, conf.Accounts.ByName("admin-account"))
	assert.Equal(t, conf.Accounts.ByName("admin-account").Address.String(), "f1d6e0586b0a20c7")
	assert.Equal(t, conf.Accounts.ByName("service-account").Address.String(), "f8d6e0586b0a20c7")
}

func Test_FromFileAccountComplex(t *testing.T) {
	b := []byte(`{
		"accounts": {
			"service-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "21c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
			},
			"admin-account-1": { "fromFile": "private.json" },
			"admin-account-3": { "fromFile": "private.json" },
			"admin-account-5": { "fromFile": "private.testnet.json" }
		}
	}`)

	b2 := []byte(`{
		 "accounts":{
				"admin-account-1":{
					 "address":"f1d6e0586b0a20c7",
					 "key":"3335dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
				},
				"admin-account-2":{
					 "address":"f2d6e0586b0a20c7",
					 "key":"3335dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
				},
				"admin-account-3":{
					 "address":"f3d6e0586b0a20c7",
					 "key":"3335dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
				}
		 }
	}`)

	b3 := []byte(`{
		 "accounts":{
				"admin-account-5":{
					 "address":"f5d6e0586b0a20c7",
					 "key":"3335dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7"
				}
		 }
	}`)

	mockFS := afero.NewMemMapFs()
	err := afero.WriteFile(mockFS, "flow.json", b, 0644)
	err2 := afero.WriteFile(mockFS, "private.json", b2, 0644)
	err3 := afero.WriteFile(mockFS, "private.testnet.json", b3, 0644)

	assert.NoError(t, err)
	assert.NoError(t, err2)
	assert.NoError(t, err3)

	composer := config.NewLoader(afero.Afero{Fs: mockFS})
	composer.AddConfigParser(json.NewParser())

	conf, loadErr := composer.Load([]string{"flow.json"})

	assert.NoError(t, loadErr)
	assert.Equal(t, 4, len(conf.Accounts))
	assert.NotNil(t, conf.Accounts.ByName("service-account"))
	assert.NotNil(t, conf.Accounts.ByName("admin-account-1"))
	assert.NotNil(t, conf.Accounts.ByName("admin-account-3"))
	assert.NotNil(t, conf.Accounts.ByName("admin-account-5"))
	assert.Equal(t, conf.Accounts.ByName("service-account").Address.String(), "f8d6e0586b0a20c7")
	assert.Equal(t, conf.Accounts.ByName("admin-account-1").Address.String(), "f1d6e0586b0a20c7")
	assert.Equal(t, conf.Accounts.ByName("admin-account-3").Address.String(), "f3d6e0586b0a20c7")
	assert.Equal(t, conf.Accounts.ByName("admin-account-5").Address.String(), "f5d6e0586b0a20c7")
}

func Test_JSONEnv(t *testing.T) {
	b := []byte(`{
		"accounts": {
			"emulator-account": {
				"address": "f8d6e0586b0a20c7",
				"key": "$EMULATOR_KEY"
			}
		}
	}`)

	os.Setenv("EMULATOR_KEY", "21c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7")

	mockFS := afero.NewMemMapFs()
	err := afero.WriteFile(mockFS, "test2-flow.json", b, 0644)

	assert.NoError(t, err)

	composer := config.NewLoader(afero.Afero{Fs: mockFS})
	composer.AddConfigParser(json.NewParser())
	conf, loadErr := composer.Load([]string{"test2-flow.json"})

	assert.NoError(t, loadErr)
	assert.Equal(t, 1, len(conf.Accounts))
	assert.Equal(t, "0x21c5dfdeb0ff03a7a73ef39788563b62c89adea67bbb21ab95e5f710bd1d40b7", conf.Accounts[0].Key.PrivateKey.String())
}
