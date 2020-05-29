# go-mime

MIME detector for human in Golang, no `/etc/mime.types` any more.

## Usage

```go
import (
    "github.com/qingstor/go-mime"
)

func main()  {
    // Get mime type via file extension.
    mimeType := mime.DetectFileExt("pdf")
    // Get mime type via file path or name.
    mimeType := mime.DetectFilePath("/srv/http/a.pdf")
}
```

---

> Built by QingStor Team.