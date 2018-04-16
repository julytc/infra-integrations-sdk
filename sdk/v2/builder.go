package v2

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"sync"

	"github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/cache"
	"github.com/pkg/errors"
)

const protocolVersion = "2"

// IntegrationBuilder provides a fluent interface for creating a configured Integration.
type IntegrationBuilder interface {
	// Build returns the Integration resulting from the applied configuration on this builder.
	// The integration data is empty, ready to add new entities' data.
	Build() (*Integration, error)
	// ParsedArguments sets the destination struct (pointer) where the command-line flags will be parsed to.
	ParsedArguments(interface{}) IntegrationBuilder
	// Synchronized sets the built Integration ready to be managed concurrently from multiple threads.
	// By default, the build integration is not synchronized.
	Synchronized() IntegrationBuilder
	// Writer sets the output stream where the integration resulting payload will be written to.
	// By default, the standard output (os.Stdout).
	Writer(io.Writer) IntegrationBuilder
	// Cache sets the Cache implementation that will be used to persist data between executions of the same integration.
	// By default, it will be a Disk-backed cache named stored in the file returned by the
	// cache.DefaultPath(integrationName) function.
	Cache(cache.Cache) IntegrationBuilder
	// NoCache disables the cache for this integration.
	NoCache() IntegrationBuilder
}

type integrationBuilderImpl struct {
	integration *Integration
	hasCache    bool
	arguments   interface{}
}

type disabledLocker struct{}

func (disabledLocker) Lock()   {}
func (disabledLocker) Unlock() {}

// NewIntegration creates a new IntegrationBuilder for the given integration name and version.
func NewIntegration(name string, version string) IntegrationBuilder {
	return &integrationBuilderImpl{
		integration: &Integration{
			Name:               name,
			ProtocolVersion:    protocolVersion,
			IntegrationVersion: version,
			Data:               []*EntityData{},
			writer:             os.Stdout, // defaults to stdout
		},
		hasCache: true,
	}
}

func (b *integrationBuilderImpl) Synchronized() IntegrationBuilder {
	b.integration.locker = &sync.Mutex{}
	return b
}

func (b *integrationBuilderImpl) Writer(writer io.Writer) IntegrationBuilder {
	b.integration.writer = writer
	return b
}

func (b *integrationBuilderImpl) ParsedArguments(dstPointer interface{}) IntegrationBuilder {
	b.arguments = dstPointer
	return b
}

func (b *integrationBuilderImpl) Cache(c cache.Cache) IntegrationBuilder {
	b.integration.Cache = c
	b.hasCache = true
	return b
}

func (b *integrationBuilderImpl) NoCache() IntegrationBuilder {
	b.integration.Cache = nil
	b.hasCache = false
	return b
}

func (b *integrationBuilderImpl) Build() (*Integration, error) {
	// Checking errors
	if b.integration.writer == nil {
		return nil, errors.New("integration writer can't be nil")
	}

	// Setting default values
	if b.integration.locker == nil {
		b.integration.locker = disabledLocker{}
	}

	// Checking arguments
	err := b.checkArguments()
	if err != nil {
		return nil, err
	}
	err = args.SetupArgs(b.arguments)
	if err != nil {
		return nil, err
	}
	defaultArgs := args.GetDefaultArgs(b.arguments)

	cache.SetupLogging(defaultArgs.Verbose)

	if b.integration.Cache == nil && b.hasCache {
		// TODO: set Log(log) function to this builder
		b.integration.Cache, err = cache.NewCache(cache.DefaultPath(b.integration.Name), cache.GlobalLog)
		if err != nil {
			return nil, fmt.Errorf("can't create cache: %s", err.Error())
		}
	}

	b.integration.prettyOutput = defaultArgs.Pretty

	return b.integration, nil
}

// Returns error if the parsed arguments destination is not from an acceptable type. It can be nil or a pointer to a
// struct.
func (b *integrationBuilderImpl) checkArguments() error {
	if b.arguments == nil {
		b.arguments = new(struct{})
		return nil
	}
	val := reflect.ValueOf(b.arguments)

	if val.Kind() == reflect.Ptr && val.Elem().Kind() == reflect.Struct {
		return nil
	}
	return errors.New("arguments must be a pointer to a struct (or nil)")
}
