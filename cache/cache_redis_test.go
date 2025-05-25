package cache

import "testing"

func Test_parseEnvVars(t *testing.T) {
	tests := []struct {
		name         string
		env          map[string]string
		wantHost     string
		wantPort     string
		wantUser     string
		wantPassword string
		wantDb       int
	}{
		{
			name: "insecure addr",
			env: map[string]string{
				"REDIS_ADDR": "127.0.0.2:1234",
			},
			wantHost: "127.0.0.2",
			wantPort: "1234",
			wantDb:   defaultRedisDB,
		},
		{
			name: "insecure hostport",
			env: map[string]string{
				"REDIS_HOST": "localhost",
				"REDIS_PORT": "1234",
			},
			wantHost: "localhost",
			wantPort: "1234",
			wantDb:   defaultRedisDB,
		},
		{
			name: "prefer addr",
			env: map[string]string{
				"REDIS_HOST": "localhost",
				"REDIS_PORT": "1234",
				"REDIS_ADDR": "127.0.0.2:5678",
			},
			wantHost: "127.0.0.2",
			wantPort: "5678",
			wantDb:   defaultRedisDB,
		},
		{
			name: "secure password",
			env: map[string]string{
				"REDIS_HOST":     "localhost",
				"REDIS_PORT":     "1234",
				"REDIS_USER":     "foo",
				"REDIS_PASSWORD": "bar",
			},
			wantHost:     "localhost",
			wantPort:     "1234",
			wantUser:     "foo",
			wantPassword: "bar",
			wantDb:       0,
		},
		{
			name: "secure pass",
			env: map[string]string{
				"REDIS_HOST": "localhost",
				"REDIS_PORT": "1234",
				"REDIS_USER": "foo",
				"REDIS_PASS": "bar",
			},
			wantHost:     "localhost",
			wantPort:     "1234",
			wantUser:     "foo",
			wantPassword: "bar",
			wantDb:       0,
		},
		{
			name: "custom db",
			env: map[string]string{
				"REDIS_HOST": "localhost",
				"REDIS_PORT": "1234",
				"REDIS_USER": "foo",
				"REDIS_PASS": "bar",
				"REDIS_DB":   "1",
			},
			wantHost:     "localhost",
			wantPort:     "1234",
			wantUser:     "foo",
			wantPassword: "bar",
			wantDb:       1,
		},
		{
			name: "host without port",
			env: map[string]string{
				"REDIS_HOST": "localhost",
			},
			wantHost: "localhost",
			wantPort: defaultRedisPort,
			wantDb:   defaultRedisDB,
		},
		{
			name: "addr without port",
			env: map[string]string{
				"REDIS_ADDR": "localhost",
			},
			wantHost: "localhost",
			wantPort: defaultRedisPort,
			wantDb:   defaultRedisDB,
		},
		{
			name: "malformed db",
			env: map[string]string{
				"REDIS_HOST": "localhost",
				"REDIS_PORT": "1234",
				"REDIS_DB":   "baz",
			},
			wantHost: "localhost",
			wantPort: "1234",
			wantDb:   defaultRedisDB,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, val := range tt.env {
				t.Setenv(key, val)
			}

			gotHost, gotPort, gotUser, gotPassword, gotDb := parseEnvVars()
			if gotHost != tt.wantHost {
				t.Errorf("parseEnvVars() gotHost = %v, want %v", gotHost, tt.wantHost)
			}
			if gotPort != tt.wantPort {
				t.Errorf("parseEnvVars() gotPort = %v, want %v", gotPort, tt.wantPort)
			}
			if gotUser != tt.wantUser {
				t.Errorf("parseEnvVars() gotUser = %v, want %v", gotUser, tt.wantUser)
			}
			if gotPassword != tt.wantPassword {
				t.Errorf("parseEnvVars() gotPassword = %v, want %v", gotPassword, tt.wantPassword)
			}
			if gotDb != tt.wantDb {
				t.Errorf("parseEnvVars() gotDb = %v, want %v", gotDb, tt.wantDb)
			}
		})
	}
}
