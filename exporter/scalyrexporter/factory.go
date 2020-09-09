// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scalyrexporter

import (
	"os"
	"fmt"
        "context"
	"errors"
	"time"

	"go.uber.org/zap"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configerror"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/exporter/exporterhelper"

)

const (
	// The value of "type" key in configuration.
	typeStr = "scalyr"

	defaultTimeout = time.Second * 5

	defaultFormat = "json"

	defaultEndpoint string = "https://app.scalyr.com"
)

func NewFactory() component.ExporterFactory {
	fmt.Printf("hello")
	return exporterhelper.NewFactory(
		typeStr,
		createDefaultConfig,
		exporterhelper.WithTraces(createTraceExporter))
}

// CreateDefaultConfig creates the default configuration for exporter.
func createDefaultConfig() configmodels.Exporter {
	endpoint := os.Getenv("SCALYR_ENDPOINT")
	if endpoint == "" {
		endpoint = "app.scalyr.com"
	}
	return &Config{
		ExporterSettings: configmodels.ExporterSettings{
			TypeVal: typeStr,
			NameVal: typeStr,
		},
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Timeout: defaultTimeout,
		},
		APIKey: os.Getenv("SCALYR_API_TOKEN"),
		Endpoint: endpoint,
	}
}

func createTraceExporter(
	_ context.Context,
	params component.ExporterCreateParams,
	cfg configmodels.Exporter,
) (component.TraceExporter, error) {
	c := cfg.(*Config)

	if c.Endpoint == "" {
		return nil, errors.New("exporter config requires a non-empty 'endpoint'")
	}

	return newTraceExporter(c, params)
}

// TODO: We should support this
func createMetricsExporter(_ *zap.Logger, _ configmodels.Exporter) (component.MetricsExporter, error) {
	return nil, configerror.ErrDataTypeIsNotSupported
}
