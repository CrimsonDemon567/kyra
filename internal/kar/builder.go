package kar

import (
    "archive/zip"
    "os"
    "path/filepath"
)

// Build creates a .kar archive from a project folder.
func Build(project string) error {
    out, _ := os.Create(project + ".kar")
    zw := zip.NewWriter(out)

    filepath.Walk(project, func(path string, info os.FileInfo, err error) error {
        if info.IsDir() {
            return nil
        }

        rel, _ := filepath.Rel(project, path)
        w, _ := zw.Create(rel)
        data, _ := os.ReadFile(path)
        w.Write(data)
        return nil
    })

    zw.Close()
    return nil
}
