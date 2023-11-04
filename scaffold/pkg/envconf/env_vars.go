package envconf

import (
	"bytes"
	"encoding"
	"encoding/base64"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

func EnvVarsFromEnviron(prefix string, envs []string) *EnvVars {
	var cryptor *AESCryptor
	// aes/ofb base64(key)
	valueEncryptedBy := os.Getenv("VALUE_ENCRYPTED_BY")

	if valueEncryptedBy != "" {
		c, err := NewAESCryptorFromKey(valueEncryptedBy)
		if err != nil {
			panic(err)
		}
		cryptor = c
	}

	e := NewEnvVars(strings.ToUpper(prefix))

	for _, kv := range envs {
		keyValuePair := strings.SplitN(kv, "=", 2)
		if len(keyValuePair) == 2 {
			k := strings.ToUpper(keyValuePair[0])

			if strings.HasPrefix(k, e.Prefix) {
				envVar := &EnvVar{KeyPath: strings.Replace(k, e.Prefix+"_", "", 1)}

				if cryptor != nil {
					v, err := base64.RawStdEncoding.DecodeString(keyValuePair[1])
					if err != nil {
						panic(err)
					}
					r, err := cryptor.DecryptFor(bytes.NewBuffer(v))
					if err != nil {
						panic(err)
					}
					data, _ := ioutil.ReadAll(r)
					envVar.Value = string(data)
				} else {
					envVar.Value = keyValuePair[1]
				}

				e.Set(envVar)
			}
		}
	}
	return e
}

func NewEnvVars(prefix string) *EnvVars {
	e := &EnvVars{
		Prefix: prefix,
	}
	return e
}

type EnvVars struct {
	Prefix string
	Values map[string]*EnvVar
}

func (e *EnvVars) Set(envVar *EnvVar) {
	if e.Values == nil {
		e.Values = map[string]*EnvVar{}
	}
	e.Values[strings.ToUpper(envVar.KeyPath)] = envVar
}

var interfaceTextMarshaller = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
var interfaceTextUnmarshaller = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

func (e *EnvVars) Len(key string) int {
	maxIdx := -1

	for _, envVar := range e.Values {
		keyPath := strings.ToUpper(envVar.KeyPath)
		k := strings.ToUpper(key)

		if strings.HasPrefix(keyPath, k) {
			v := strings.TrimLeft(keyPath, k+"_")
			parts := strings.Split(v, "_")
			i, err := strconv.ParseInt(parts[0], 10, 64)
			if err == nil {
				if int(i) > maxIdx {
					maxIdx = int(i)
				}
			}
		}
	}

	return maxIdx + 1
}

func (e *EnvVars) Get(key string) *EnvVar {
	if e.Values == nil {
		return nil
	}
	return e.Values[strings.ToUpper(key)]
}

func (e *EnvVars) MaskBytes() []byte {
	return e.DotEnv(func(envVar *EnvVar) string {
		if envVar.Mask != "" {
			return envVar.Mask
		}
		return envVar.Value
	})
}

func (e *EnvVars) Bytes() []byte {
	return e.DotEnv(func(envVar *EnvVar) string {
		return envVar.Value
	})
}

func (e *EnvVars) DotEnv(valuer func(envVar *EnvVar) string) []byte {
	values := map[string]string{}
	for _, envVar := range e.Values {
		values[envVar.Key(e.Prefix)] = valuer(envVar)
	}
	return DotEnv(values)
}

func DotEnv(keyValues map[string]string) []byte {
	buf := bytes.NewBuffer(nil)

	keys := make([]string, 0)
	for k := range keyValues {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		buf.WriteString(k)
		buf.WriteRune('=')
		buf.WriteString(keyValues[k])
		buf.WriteRune('\n')
	}

	return buf.Bytes()
}
