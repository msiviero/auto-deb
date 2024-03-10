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
	version := flag.String("v", "0.0-dev", "version")
	flag.Parse()

	cfg := parseConfig(*configfile)

	cfg.Service.Environment["APP_VERSION"] = *version

	if len(*version) > 0 {
		cfg.Package.Version = *version
	}

	pkgname := fmt.Sprintf("%s-%s", cfg.Package.Name, cfg.Package.Version)

	ensureDirTree(*outdir, pkgname)
	debianControl(*outdir, pkgname, cfg)
	debianPreinst(*outdir, pkgname, cfg)
	debianPostinst(*outdir, pkgname, cfg)
	debianPrerm(*outdir, pkgname, cfg)
	systemdService(*outdir, pkgname, cfg)
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
	User        string            `yaml:"user"`
	Environment map[string]string `yaml:"environment"`
}

func ensureDirTree(outdir string, pkgname string) {

	must(os.MkdirAll(filepath.Join(
		outdir,
		pkgname,
		"DEBIAN",
	), os.ModePerm))

	must(os.MkdirAll(filepath.Join(
		outdir,
		pkgname,
		"etc",
		"systemd",
		"system",
	), os.ModePerm))

	must(os.MkdirAll(filepath.Join(
		outdir,
		pkgname,
		"usr",
		"local",
		"bin",
	), os.ModePerm))
}

func debianControl(outdir string, pkgname string, cfg DebianConf) {
	filehandler := check(os.OpenFile(
		filepath.Join(outdir, pkgname, "DEBIAN", "control"),
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

func debianPreinst(outdir string, pkgname string, cfg DebianConf) {
	filehandler := check(os.OpenFile(
		filepath.Join(outdir, pkgname, "DEBIAN", "preinst"),
		os.O_WRONLY|os.O_CREATE,
		0775,
	))
	defer filehandler.Close()

	tmpl := check(template.New("preinst").Parse(preinstTpl))
	must(tmpl.Execute(filehandler, map[string]string{
		"name": cfg.Package.Name,
	}))
}

func debianPostinst(outdir string, pkgname string, cfg DebianConf) {
	filehandler := check(os.OpenFile(
		filepath.Join(outdir, pkgname, "DEBIAN", "postinst"),
		os.O_WRONLY|os.O_CREATE,
		0775,
	))
	defer filehandler.Close()

	tmpl := check(template.New("postinst").Parse(postinstTpl))
	must(tmpl.Execute(filehandler, map[string]string{
		"name": cfg.Package.Name,
	}))
}

func debianPrerm(outdir string, pkgname string, cfg DebianConf) {
	filehandler := check(os.OpenFile(
		filepath.Join(outdir, pkgname, "DEBIAN", "prerm"),
		os.O_WRONLY|os.O_CREATE,
		0775,
	))
	defer filehandler.Close()

	tmpl := check(template.New("prerm").Parse(prermTpl))
	must(tmpl.Execute(filehandler, map[string]string{
		"name": cfg.Package.Name,
	}))
}

func systemdService(outdir string, pkgname string, cfg DebianConf) {
	filehandler := check(os.OpenFile(
		filepath.Join(
			outdir,
			pkgname,
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
		"workingdir":  filepath.Join("/home", cfg.Service.User),
		"user":        cfg.Service.User,
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
