package types

import (
	"errors"
	"fmt"
	"github.com/justinbarrick/farm/pkg/cache/file"
	"github.com/justinbarrick/farm/pkg/cache/s3"
	"github.com/justinbarrick/farm/pkg/job"
	"github.com/justinbarrick/farm/pkg/executors/kubernetes"
)

type Config struct {
	Jobs   []*job.Job   `hcl:"job,block"`
	Cache  *CacheConfig `hcl:"cache,block"`
	Kubernetes *kubernetes.Kubernetes `hcl:"kubernetes,block"`
	Engine *string      `hcl:"engine"`
}

type CacheConfig struct {
	S3   *s3cache.S3Cache     `hcl:"s3,block"`
	File *filecache.FileCache `hcl:"file,block"`
}

func (c Config) Validate() error {
	for _, job := range c.Jobs {
		if err := job.Validate(c.GetEngine()); err != nil {
			return errors.New(fmt.Sprintf("Error validating job %s: %s", job.Name, err))
		}
	}

	return nil
}

func (c Config) GetEngine() string {
	if c.Engine != nil {
		return *c.Engine
	}
	return ""
}
