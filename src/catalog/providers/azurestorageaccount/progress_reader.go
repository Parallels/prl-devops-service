package azurestorageaccount

// import (
// 	"fmt"
// 	"io"

// 	"github.com/Parallels/pd-api-service/basecontext"
// )

// type progressReader struct {
// 	op       string
// 	r        io.Reader
// 	progress int64
// 	total    int64
// 	filename string
// 	ctx      basecontext.ApiContext
// }

// func (pr *progressReader) Read(p []byte) (int, error) {
// 	n, err := pr.r.Read(p)
// 	pr.progress += int64(n)
// 	if pr.total > 0 {
// 		pr.ctx.LogInfof("%s %v of %v from %s", pr.op, pr.getFormattedProgress(pr.progress), pr.getFormattedProgress(pr.total), pr.filename)
// 	} else {
// 		pr.ctx.LogInfof("%s %v MBs from %s", pr.op, pr.getFormattedProgress(pr.progress), pr.filename)

// 	}
// 	return n, err
// }

// func (pr *progressReader) getFormattedProgress(val int64) string {
// 	if val > 0 && val < 1024 {
// 		return fmt.Sprintf("%.2f", float64(val)) + " bytes"
// 	} else if val > 0 && val < 1024*1024 {
// 		kbs := float64(val) / 1024
// 		return fmt.Sprintf("%.2f", kbs) + " KBs"
// 	} else if val > 0 && val < 1024*1024*1024 {
// 		mbs := float64(val) / (1024 * 1024)
// 		return fmt.Sprintf("%.2f", mbs) + " MBs"
// 	} else if val > 0 && val < 1024*1024*1024*1024 {
// 		gbs := float64(val) / (1024 * 1024 * 1024)
// 		return fmt.Sprintf("%.2f", gbs) + " GBs"
// 	} else {
// 		rest := float64(val) / (1024 * 1024 * 1024 * 1024)
// 		return fmt.Sprintf("%.2f", rest) + " TBs"
// 	}
// }
