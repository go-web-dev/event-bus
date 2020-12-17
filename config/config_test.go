package config

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type configSuite struct {
	suite.Suite
}

func TestConfig(t *testing.T) {
	suite.Run(t, new(configSuite))
}
