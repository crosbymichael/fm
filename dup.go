package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/shlex"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

var dupsCommand = cli.Command{
	Name:  "d",
	Usage: "fine duplicates",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "hidden",
			Usage: "show hidden files",
		},
	},
	Action: func(clix *cli.Context) (err error) {
		path := clix.Args().First()
		if path == "" {
			if path, err = filepath.Abs("."); err != nil {
				return errors.Wrap(err, "cannot get absolute path")
			}
		}
		w := &walker{
			base:     path,
			hash:     true,
			handlers: []filepath.WalkFunc{skipDirs, skipPermErr},
		}
		if !clix.Bool("hidden") {
			w.handlers = append(w.handlers, skipHidden)
		}
		if err := filepath.Walk(path, w.walk); err != nil {
			return errors.Wrapf(err, "walking %s", path)
		}
		m := make(map[string][]*extInfo)
		for _, r := range w.results {
			k := r.MD5
			if r.Size() == 0 {
				k = "empty"
			}
			m[k] = append(m[k], r)
		}
		tmp, err := tempFileDups(path, m)
		if err != nil {
			return err
		}
		defer os.Remove(tmp)

		if err := startEditor(tmp); err != nil {
			return errors.Wrap(err, "running editor")
		}
		dm, err := createDupMap(tmp)
		if err != nil {
			return err
		}
		for k, vv := range dm {
			fmt.Println(k, vv)
		}
		return nil
	},
}

func tempFileDups(path string, results map[string][]*extInfo) (string, error) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		return "", errors.Wrap(err, "create tmp file")
	}
	defer f.Close()

	for k, items := range results {
		if len(items) > 1 {
			fmt.Fprintf(f, "# %s\n", k)
			for _, r := range items {
				rel, err := display(path, r.Path)
				if err != nil {
					return "", errors.Wrap(err, "get display path")
				}
				if _, err := fmt.Fprintf(f, "%s\n", rel); err != nil {
					return "", errors.Wrap(err, "write path to file")
				}
			}
		}
	}
	return f.Name(), nil
}

type Command int

const (
	rm Command = iota + 1
	mv
)

func createDupMap(path string) (map[Command][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "open %s", path)
	}
	defer f.Close()

	var (
		out = make(map[Command][]string)
		s   = bufio.NewScanner(f)
	)
	for s.Scan() {
		if err := s.Err(); err != nil {
			return nil, errors.Wrap(err, "scan error")
		}
		line := strings.TrimSpace(s.Text())
		if line == "" {
			return nil, errors.New("unable to continue with empty line")
		}
		if line[0] == skipMoveToken {
			continue
		}
		parts, err := shlex.Split(line)
		if err != nil {
			return nil, err
		}
		switch parts[0] {
		case "rm", "d":
			out[rm] = []string{
				parts[1],
			}
		case "mv", "m":
			out[mv] = []string{
				parts[1],
				parts[2],
			}
		}
	}
	return out, nil
}
