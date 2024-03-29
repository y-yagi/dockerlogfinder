package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/olekukonko/tablewriter"
)

func main() {
	os.Exit(run(os.Args))
}

func msg(err error) int {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err)
		return 1
	}
	return 0
}

func run(args []string) int {
	cli, err := client.NewClientWithOpts(client.WithVersion("1.40"))
	if err != nil {
		return msg(err)
	}

	volumes, err := cli.VolumeList(context.Background(), filters.NewArgs())
	if err != nil {
		return msg(err)
	}

	if args[1] == "--list" {
		showList(volumes)
		return 0
	}

	var wg sync.WaitGroup
	for _, volume := range volumes.Volumes {
		wg.Add(1)
		go func(location string) {
			search(location, args[1:], os.Stdout)
			wg.Done()
		}(volume.Mountpoint)
	}
	wg.Wait()

	return 0
}

func search(location string, names []string, outStream io.Writer) {
	err := filepath.Walk(location, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		for _, name := range names {
			matched, _ := filepath.Match(name, info.Name())
			if matched {
				fmt.Fprintf(outStream, "%s\n", path)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(outStream, "%v\n", err)
	}
}

func showList(volumes volume.ListResponse) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Driver", "Name", "Mountpoint"})

	for _, volume := range volumes.Volumes {
		table.Append([]string{volume.Driver, volume.Name, volume.Mountpoint})
	}

	table.Render()
}
