package upload

import (
	"fmt"
	"io"
	"os/exec"
	"path"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe/manifest"
)

type Targets struct {
	Targets map[string]interface{}
}

type Plan []Command

func (p Plan) Execute(stdout io.Writer, stdin io.Reader) (err error) {
	fmt.Fprintln(stdout, "Planned execution") // #nosec
	for _, cmd := range p {
		fmt.Fprintf(stdout, "\t* %s\n", cmd) // #nosec
	}
	fmt.Fprint(stdout, "\nAre you sure? [y/N]: ") // #nosec
	var input string
	fmt.Fscan(stdin, &input) // #nosec
	if input != "y" {
		err = errors.New("aborted")
	}

	for _, cmd := range p {
		fmt.Fprintf(stdout, "\n=== %s ===\n", cmd) // #nosec
		runErr := cmd.Cmd.Run()
		if runErr != nil {
			err = runErr
			return
		}
	}

	return
}

type Command struct {
	Cmd       exec.Cmd
	Printable string
}

func (c Command) String() string {
	if c.Printable != "" {
		return c.Printable
	}
	return strings.Join(c.Cmd.Args, " ")

}

type Planner interface {
	Plan() (plan Plan, err error)
}

func NewPlanner(fs afero.Afero, pathResolver PathResolver, homedir string, stdout io.Writer, stderr io.Writer, stdin io.Reader, pipelineFile PipelineFile) Planner {
	return planner{
		fs:           fs,
		pathResolver: pathResolver,
		homedir:      homedir,
		stdout:       stdout,
		stderr:       stderr,
		stdin:        stdin,
		pipelineFile: pipelineFile,
	}
}

type PathResolver func(string) (string, error)
type PipelineFile func(fs afero.Afero) (afero.File, error)
type planner struct {
	fs           afero.Afero
	pathResolver PathResolver
	homedir      string
	stdout       io.Writer
	stderr       io.Writer
	stdin        io.Reader
	pipelineFile PipelineFile
}

func (p planner) getHalfpipeManifest() (man manifest.Manifest, err error) {
	bytes, err := p.fs.ReadFile(".halfpipe.io")
	if err != nil {
		return
	}

	err = yaml.Unmarshal(bytes, &man)
	if err != nil {
		return
	}

	if man.Team == "" || man.Pipeline == "" {
		err = errors.New("'team' and 'pipeline' must be defined in '.halfpipe.io'")
	}

	return
}

func (p planner) getTargets() (targets Targets, err error) {
	path := path.Join(p.homedir, ".flyrc")
	exists, err := p.fs.Exists(path)
	if err != nil {
		return
	}

	if !exists {
		return
	}

	bytes, err := p.fs.ReadFile(path)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(bytes, &targets)
	return
}

func (p planner) loginCommand(team string) (cmd Command, err error) {
	path, err := p.pathResolver("fly")
	if err != nil {
		return
	}

	cmd.Cmd = exec.Cmd{
		Path:   path,
		Args:   []string{path, "-t", team, "login", "-c", "https://concourse.halfpipe.io", "-n", team},
		Stdout: p.stdout,
		Stderr: p.stderr,
		Stdin:  p.stdin,
	}

	return
}

func (p planner) lintAndRender() (cmd Command, err error) {
	file, err := p.pipelineFile(p.fs)
	if err != nil {
		return
	}

	path, err := p.pathResolver("halfpipe")
	if err != nil {
		return
	}

	cmd.Cmd = exec.Cmd{
		Path:   path,
		Args:   []string{path},
		Stderr: p.stderr,
		Stdout: file,
	}
	cmd.Printable = fmt.Sprintf("%s > %s", path, file.Name())

	return
}

func (p planner) Plan() (plan Plan, err error) {
	man, err := p.getHalfpipeManifest()
	if err != nil {
		return
	}

	targets, err := p.getTargets()
	if err != nil {
		return
	}

	if _, ok := targets.Targets[man.Team]; !ok {
		cmd, loginError := p.loginCommand(man.Team)
		if loginError != nil {
			err = loginError
			return
		}

		plan = append(plan, cmd)
	}

	cmd, err := p.lintAndRender()
	if err != nil {
		return
	}
	plan = append(plan, cmd)

	return
}
