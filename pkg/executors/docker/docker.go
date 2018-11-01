package docker

import (
	"context"
	"io"
	"fmt"
	"log"
	"os"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	docker "github.com/docker/docker/client"
	"github.com/docker/docker/api/types/container"
	"github.com/justinbarrick/farm/pkg/config"
)

func Run(j config.Job) error {
	log.Printf("===> Running job: %s\n", j.Name)
	ctx := context.TODO()

	d, err := docker.NewEnvClient()
	if err != nil {
		return err
	}

	args := filters.NewArgs()
	args.Add("reference", j.Image)

	images, err := d.ImageList(ctx, types.ImageListOptions{
		Filters: args,
	})
	if err != nil {
		return err
	}

	if len(images) < 1 {
		reader, err := d.ImagePull(ctx, j.Image, types.ImagePullOptions{})
		if err != nil {
			return err
		}
		io.Copy(os.Stdout, reader)
	}

	env := []string{}
	for name, value := range j.Env {
		env = append(env, fmt.Sprintf("%s=%s", name, value))
	}

	ctr, err := d.ContainerCreate(ctx, &container.Config{
		Image: j.Image,
		Cmd: []string{
			j.Shell,
		},
		Entrypoint: []string{"/bin/sh", "-c"},
		Env: env,
	}, nil, nil, "")
	if err != nil {
		return err
	}

	if err := d.ContainerStart(ctx, ctr.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	log.Printf("Started container: %s\n", ctr.ID[:8])

	_, err = d.ContainerWait(ctx, ctr.ID)
	if err != nil {
		return err
	}

	out, err := d.ContainerLogs(ctx, ctr.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, out)
	log.Printf("===> Job completed: %s\n", j.Name)
	return nil
}
