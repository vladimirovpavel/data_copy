package main

//remember about BUFIO

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/cheggaaa/pb/v3"
)

const bufSize = 1024

//struct for file instance. Contains path, len, handle to file
type CopyFile struct {
	path   string
	len    uint64
	handle *os.File
}

// open file for read, check it size, compare with requested count. Saves handle to c.handle
func (c *CopyFile) checkRead(offset uint64, count *uint64) error {
	handle, err := os.OpenFile(c.path, os.O_RDONLY, 0222)
	if err != nil {
		err = fmt.Errorf("error on file %s for read %v", c.path, err)
		return err
	}
	fi, err := handle.Stat()
	if err != nil {
		err = fmt.Errorf("error in getting stat of file")
		return err
	}
	c.handle = handle
	c.len = uint64(fi.Size())

	if offset > c.len {
		err = fmt.Errorf("wants read from offset %d, but file len is %d", offset, c.len)
		return err
	}

	if *count == 0 {
		*count = c.len - offset
		fmt.Printf("Count is 0. Setted it to filesize - offset = %d\n", *count)
	}
	c.handle.Seek(int64(offset), io.SeekStart)
	return nil
}

// Open file to write, saves handle to c.handle. Throw err if error
func (c *CopyFile) checkWrite() error {
	handle, err := os.OpenFile(c.path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		err = fmt.Errorf("error on open file %s for write %v", c.path, err)
		return err
	}
	c.handle = handle
	return nil
}

//Works func. Creates buf for read and write.
//In cycle read data from source, write to Dest
//If progress + bufSize > limit (buf more than bytes to read) - creates new buf to read data.
//In end of each iterationg throw progress to chan (to ProgressPrint)
func copy(sourceCopyFile *CopyFile, destCopyFile *CopyFile, limit uint64) error {
	progress := 0
	buf := make([]byte, bufSize)
	bar := pb.StartNew(int(limit))

	for progress < int(limit) {
		//ERRORS, EOF
		if left := progress + bufSize; left > int(limit) {
			buf = make([]byte, uint32(limit)-uint32(progress))
		}
		readed, err := sourceCopyFile.handle.Read(buf)
		if err != nil {
			return err
		}
		writed, err := destCopyFile.handle.Write(buf)

		if err != nil {
			return err
		} else if writed != readed {
			return fmt.Errorf("error, writed %d bytes, less then readed %d", writed, readed)
		}
		progress += readed
		bar.Add(bufSize)
	}
	bar.Finish()
	return nil

}

//Func receives source, dest, offset to read from source, count bytes to write
//creates the CopyFile structs, checks correctly. (File handlers close - defer call)
//reduces byte to write if Count less then data in source file
//
//work copy func starts in other gorutine
//ProgressPrint starts in other gorutine
//created WaitGroup, waits after copy is ending
func Copy(source string, dest string, offset uint64, count uint64) error {
	sourceCopyFile := &CopyFile{path: source}
	destCopyFile := &CopyFile{path: dest}

	from_err := sourceCopyFile.checkRead(offset, &count)
	if from_err != nil {
		return from_err
	}

	to_err := destCopyFile.checkWrite()
	if to_err != nil {
		return to_err
	}
	defer sourceCopyFile.handle.Close()
	defer destCopyFile.handle.Close()

	if summ := offset + count; summ > sourceCopyFile.len {
		fmt.Printf("warning: count of bytes off the source file boundaries by %d. Read to EOF\n", summ-sourceCopyFile.len)
		count = sourceCopyFile.len - offset
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func(from_cf *CopyFile, to_cf *CopyFile, count uint64) {
		defer wg.Done()
		copy(from_cf, to_cf, count)
	}(sourceCopyFile, destCopyFile, count)

	wg.Wait()

	return nil
}
