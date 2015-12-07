package link8

type linker struct {
	pkgs map[string]*Pkg
}

func newLinker() *linker {
	ret := new(linker)
	ret.pkgs = make(map[string]*Pkg)
	return ret
}

func (lnk *linker) addPkg(p *Pkg) bool {
	path := p.path
	if _, found := lnk.pkgs[path]; found {
		return false
	}

	lnk.pkgs[path] = p
	return true
}

// addPkgs add a package and recursively adds
// all the packages that this package imported.
func (lnk *linker) addPkgs(p *Pkg) {
	if lnk.addPkg(p) {
		for _, req := range p.imported {
			lnk.addPkgs(req)
		}
	}
}

func (lnk *linker) pkg(path string) *Pkg {
	ret, found := lnk.pkgs[path]
	if !found {
		panic("package missing")
	}
	return ret
}
