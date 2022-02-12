package main

//CHECKS TYPES OF INT/UINT
import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var TESTFILENAME string = "c:\\temp\\test_file1.txt"
var TESTFILENAME2 string = "c:\\temp\\test_file2.txt"
var LINUXTESTFILENAME string = "/home/user/golang/tests/test_file1.txt"
var LINUXTESTFILENAME2 string = "/home/user/golang/tests/test_file2.txt"
var TESTFILESIZE = 500
var TESTFILESIZE2 = 1024 * 1024

func createTestFile(name string, len int32) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()
	buf := make([]byte, len)
	for i := range buf {
		buf[i] = byte(i) % 255
	}
	_, err = f.Write(buf)
	if err != nil {
		return err
	}
	return nil
}

func TestCheckRead(t *testing.T) {
	var offset, count uint64
	// EMPTY FILENAME
	emptyNameCF := CopyFile{path: ""}
	require.Error(t, emptyNameCF.checkRead(offset, &count), "Empty filename")
	//NOT EXIST FILENAME
	notExistingCF := CopyFile{path: "c:\\temp\\not_existing_file.txt"}
	require.Error(t, notExistingCF.checkRead(offset, &count), "Not existing filename")

	err := createTestFile(LINUXTESTFILENAME, int32(TESTFILESIZE))
	if err != nil {
		t.Fatalf("Error on creating test file: :%v", err)
	}
	fr := CopyFile{path: LINUXTESTFILENAME}
	defer fr.handle.Close()

	tests := map[struct{ offset, count uint64 }]bool{
		{0, 10}:    true,  //normal reading
		{100, 600}: true,  //read from offset, count > file(len) - offset
		{100, 0}:   true,  //zero count (get all file size)
		{55, 1000}: true,  //has offset and wery big count
		{600, 10}:  false, //offset > len(file)
		{1000, 50}: false, //offset > len(file)
	}
	for test_case, result := range tests {
		if !result {
			require.Error(t, fr.checkRead(test_case.offset, &test_case.count), "Testing offset off the input file")
		} else {
			require.Nil(t, fr.checkRead(test_case.offset, &test_case.count), "Testing offset %d, count %d", offset, count)
		}
	}
	//checks count == len(file) - offset if count == 0

	offset, count = 10, 0
	fr.checkRead(offset, &count)
	require.Equal(t, count, uint64(TESTFILESIZE-int(offset)), "Checks that count == 0 is setted to FILESIZE - offset")

}

func TestCheckWrite(t *testing.T) {

	emptyFilenameCF := CopyFile{path: ""}
	require.Error(t, emptyFilenameCF.checkWrite(), "Empty filename")

	notExistingCF := CopyFile{path: "c:\\temp\\not_existing_file.txt"}
	require.Nil(t, notExistingCF.checkWrite(), "Not existing filename")
	notExistingCF.handle.Close()
	os.Remove("c:\\temp\\not_existing_file.txt")

	fw := CopyFile{path: LINUXTESTFILENAME}
	require.Nil(t, fw.checkWrite(), "Testing open file for write")
	fw.handle.Close()

	//fw = CopyFile{path: "c:\\file.txt"}
	//require.Error(t, fw.checkWrite(), "Testing create file in c:\\")
	fw = CopyFile{path: "/etc/file.txt"}
	require.Error(t, fw.checkWrite(), "Testing create file in /etc/")

	fw.handle.Close()
}

func TestProgessPrint(t *testing.T) {

}

func TestCopy(t *testing.T) {
	//require.Error(t, Copy("c:\\file.txt", "c:\\file2.txt", 0, 100), "Open not existing file and try to write to forbidden dir")
	require.Error(t, Copy("/home/user/some_src_file", "/home/user/some_dst_file", 0, 100), "Open not existing file and try to write to forbidden dir")

	err := Copy("/home/user/test-file", "/home/user/dst_file", 100, 0)
	if err != nil {
		t.Fatalf("Error copying big file")
	}

	err = createTestFile(LINUXTESTFILENAME, int32(TESTFILESIZE))
	if err != nil {
		t.Fatalf("Error on creating file: %v", err)
	}
	require.Error(t, Copy(LINUXTESTFILENAME, "/etc/test.txt", 0, 10000), "Try to copy existing file to forbidden place")

	err = createTestFile(LINUXTESTFILENAME, int32(TESTFILESIZE2))
	if err != nil {
		t.Fatalf("ERror on creating temp file: %v", err)
	}
	err = Copy(LINUXTESTFILENAME, LINUXTESTFILENAME, 0, 100000000)
	if err != nil {
		t.Fatalf("Error on copy files: %v", err)
	}

	f, err := os.OpenFile(LINUXTESTFILENAME, os.O_RDONLY, 0666)
	if err != nil {
		t.Fatalf("Error open writed file for test it size")
	}
	writedFileState, err := f.Stat()
	if err != nil {
		t.Fatal("Error receive stat from writed file")
	}
	require.Equal(t, writedFileState.Size(), int64(TESTFILESIZE2), "Testing then len(writed file) == len(source file)")
	f.Close()
	os.Remove(LINUXTESTFILENAME)
	os.Remove(LINUXTESTFILENAME)
}
