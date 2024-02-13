package azurestorageaccount

// type progressWriter struct {
// 	w        io.Writer
// 	progress int64
// 	ctx      basecontext.ApiContext
// 	filename string
// }

// func (pw *progressWriter) Write(p []byte) (int, error) {
// 	n, err := pw.w.Write(p)
// 	pw.progress += int64(n)
// 	pw.ctx.LogInfof("Pulled %d bytes from %s", pw.progress, pw.filename)
// 	return n, err
// }
