package config_test

import (
	"testing"

	"github.com/samverrall/polybin/config"
	"github.com/stretchr/testify/assert"
)

func toPointer[T any](value T) *T {
	return &value
}

func TestParse(t *testing.T) {
	const (
		exampleConfig = "./testdata/example_config.json"
	)
	tests := []struct {
		name     string
		filepath string
		want     *config.Config
		wantErr  bool
	}{
		{
			name:     "Successful config parse",
			filepath: exampleConfig,
			want: &config.Config{
				config.ConfigEntry{
					ProjectName: "testproject",
					Services: []config.Service{
						{
							Type:   "watch",
							Dir:    "/testdir",
							Binary: toPointer("./test"),
							Args:   []string{"./test.sh"},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tc := range tests {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			got, err := config.Parse(tt.filepath)
			if (err != nil) != tt.wantErr {
				t.Errorf("got err: %v", err)
				t.FailNow()
			}

			assert.Equal(t, tt.want, got, "wanted: %v, got: %v", tt.want, got)
		})
	}
}
