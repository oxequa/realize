package realize

import (
	"go/build"
)

// FindGoFiles returns a list of .go source files that are
// required to build the package located in dir.
// This includes .go files of all dependencies
// except for source files from the standard library.
func FindGoFiles(dir string) ([]string, error) {
	p, err := build.ImportDir(dir, 0)
	if err != nil {
		return nil, err
	}

	// Keep track of the imports we still have to process.
	imports := p.Imports

	// Keep track of which packages we already processed
	// and which directory they were imported from.
	seen := make(map[string]string)
	for _, i := range imports {
		seen[i] = p.Dir
	}

	files := make([]string, 0)
	for _, f := range p.GoFiles {
		files = append(files, p.Dir+"/"+f)
	}

	// Keep going until we have no imports left.
	for len(imports) > 0 {
		name := imports[0]
		imports = imports[1:]

		p, err := build.Import(name, seen[name], 0)
		if err != nil {
			return nil, err
		}

		// Ignore the standard library packages.
		if p.Goroot {
			continue
		}

		for _, name := range p.Imports {
			if _, ok := seen[name]; !ok {
				seen[name] = p.Dir
				imports = append(imports, name)
			}
		}

		for _, f := range p.GoFiles {
			files = append(files, p.Dir+"/"+f)
		}
	}

	return files, nil
}
