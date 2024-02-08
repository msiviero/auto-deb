package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"gopkg.in/yaml.v3"

	_ "embed"
)

//go:embed tpl/control.tmpl
var controlTpl string

//go:embed tpl/service.tmpl
var serviceTpl string

//go:embed tpl/preinst.tmpl
var preinstTpl string

//go:embed tpl/postinst.tmpl
var postinstTpl string

//go:embed tpl/prerm.tmpl
var prermTpl string

func main() {

	configfile := flag.String("c", "./debian.yml", "config-file")
	outdir := flag.String("o", ".", "outdir")
	version := flag.String("v", "", "version")
	flag.Parse()

	cfg := parseConfig(*configfile)

	if len(*version) > 0 {
		cfg.Package.Version = *version
	}

	ensureDirTree(*outdir)
	debianControl(*outdir, cfg)
	debianPreinst(*outdir, cfg)
	debianPostinst(*outdir, cfg)
	debianPrerm(*outdir, cfg)
	systemdService(*outdir, cfg)
}

type DebianConf struct {
	Package PackageConf `yaml:"package"`
	Service ServiceConf `yaml:"service"`
}

type PackageConf struct {
	Name         string `yaml:"name"`
	Version      string `yaml:"version"`
	Architecture string `yaml:"architecture"`
	Maintainer   string `yaml:"maintainer"`
	Description  string `yaml:"description"`
}

type ServiceConf struct {
	Environment map[string]string `yaml:"environment"`
}

func ensureDirTree(outdir string) {
	must(os.MkdirAll(filepath.Join(
		outdir,
		"DEBIAN",
		"etc",
		"systemd",
		"system",
	), os.ModePerm))

	must(os.MkdirAll(filepath.Join(
		outdir,
		"DEBIAN",
		"usr",
		"local",
		"bin",
	), os.ModePerm))
}

func debianControl(outdir string, cfg DebianConf) {
	filehandler := check(os.OpenFile(
		filepath.Join(outdir, "DEBIAN", "control"),
		os.O_WRONLY|os.O_CREATE,
		0644,
	))
	defer filehandler.Close()

	tmpl := check(template.New("control").Parse(controlTpl))
	must(tmpl.Execute(filehandler, map[string]string{
		"name":         cfg.Package.Name,
		"architecture": cfg.Package.Architecture,
		"maintainer":   cfg.Package.Maintainer,
		"description":  cfg.Package.Description,
		"version":      cfg.Package.Version,
	}))
}

func debianPreinst(outdir string, cfg DebianConf) {
	filehandler := check(os.OpenFile(
		filepath.Join(outdir, "DEBIAN", "preinst"),
		os.O_WRONLY|os.O_CREATE,
		0775,
	))
	defer filehandler.Close()

	tmpl := check(template.New("preinst").Parse(preinstTpl))
	must(tmpl.Execute(filehandler, map[string]string{
		"name": cfg.Package.Name,
	}))
}

func debianPostinst(outdir string, cfg DebianConf) {
	filehandler := check(os.OpenFile(
		filepath.Join(outdir, "DEBIAN", "postinst"),
		os.O_WRONLY|os.O_CREATE,
		0775,
	))
	defer filehandler.Close()

	tmpl := check(template.New("postinst").Parse(postinstTpl))
	must(tmpl.Execute(filehandler, map[string]string{
		"name": cfg.Package.Name,
	}))
}

func debianPrerm(outdir string, cfg DebianConf) {
	filehandler := check(os.OpenFile(
		filepath.Join(outdir, "DEBIAN", "prerm"),
		os.O_WRONLY|os.O_CREATE,
		0775,
	))
	defer filehandler.Close()

	tmpl := check(template.New("prerm").Parse(prermTpl))
	must(tmpl.Execute(filehandler, map[string]string{
		"name": cfg.Package.Name,
	}))
}

func systemdService(outdir string, cfg DebianConf) {
	filehandler := check(os.OpenFile(
		filepath.Join(
			outdir,
			"DEBIAN",
			"etc", "systemd", "system",
			fmt.Sprintf("%s.service", cfg.Package.Name),
		),
		os.O_WRONLY|os.O_CREATE,
		0775,
	))
	defer filehandler.Close()

	tmpl := check(template.New("service").Parse(serviceTpl))
	must(tmpl.Execute(filehandler, map[string]any{
		"executable":  cfg.Package.Name,
		"environment": cfg.Service.Environment,
		"description": cfg.Package.Description,
		"workingdir":  filepath.Join("/home", cfg.Package.Name),
		"user":        cfg.Package.Name,
	}))
}

func parseConfig(configfile string) DebianConf {
	data := check(os.ReadFile(configfile))

	t := DebianConf{}
	must(yaml.Unmarshal([]byte(data), &t))
	return t
}

func check[T any](obj T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return obj
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
