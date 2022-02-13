package main

//CHECKS TYPES OF INT/UINT
import (
	"os"
	"os/user"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

var WIN_WRONG_FILENAME string = "c:\\wrong_file_name"
var LINUX_WRONG_FILENAME string = "/dev/wrong_file_name"

var testFiles = []struct {
	name string
	size uint64
}{
	{name: "small_test_file", size: 500},
	{name: "big_test_file", size: 1024 * 1024 * 100}}

const OS_LINUX int = 1
const OS_WINDOWS int = 2

var os_type int = OS_LINUX

func getWrongName(os_type int) string {
	wrong_name := ""
	if os_type == OS_LINUX {
		wrong_name = LINUX_WRONG_FILENAME
	} else if os_type == OS_WINDOWS {
		wrong_name = WIN_WRONG_FILENAME
	}
	return wrong_name
}

func cleanupTestEnviromnent(workDir string) error {
	err := os.RemoveAll(workDir)
	return err
}

func createTestEnviromnent() (string, error) {
	//TEST working ON WINDOWS
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	workDir := path.Join(usr.HomeDir, "tests")
	err = os.MkdirAll(workDir, 0777)
	if err != nil {
		return "", err
	}

	for _, testFile := range testFiles {
		err = createTestFile(path.Join(workDir, testFile.name), testFile.size)
		if err != nil {
			return "", err
		}
	}
	return workDir, err
}

func createTestFile(name string, len uint64) error {
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
	workDir, err := createTestEnviromnent()
	if err != nil {
		t.Fatalf("Error on create test enviromnent")
	}

	wrongName := getWrongName(os_type)
	var offset, count uint64

	emptyNameCF := structFile{path: ""}
	require.Error(t, emptyNameCF.checkRead(offset, &count), "Empty filename")

	notExistingCF := structFile{path: wrongName}
	require.Error(t, notExistingCF.checkRead(offset, &count), "Not existing filename")

	testFile := testFiles[0]
	fileRead := structFile{path: path.Join(workDir, testFile.name)}
	defer fileRead.Close()

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
			require.Error(t, fileRead.checkRead(test_case.offset, &test_case.count), "Testing offset %d off the input file", offset)
		} else {
			require.Nil(t, fileRead.checkRead(test_case.offset, &test_case.count), "Testing offset %d, count %d", offset, count)
		}
	}
	//checks count == len(file) - offset if count == 0

	offset, count = 10, 0
	fileRead.checkRead(offset, &count)
	require.Equal(t, count, uint64(uint64(testFile.size)-offset), "Checks that count == 0 is setted to FILESIZE - offset")

	cleanupTestEnviromnent(workDir)
}

func TestCheckWrite(t *testing.T) {
	workDir, err := createTestEnviromnent()
	if err != nil {
		t.Fatalf("error creating test enviromnent %v", err)
	}
	wrongName := getWrongName(os_type)
	testFile := testFiles[0]

	emptyFilenameCF := structFile{path: ""}
	require.Error(t, emptyFilenameCF.checkWrite(), "Empty filename")

	notExistingCF := structFile{path: path.Join(workDir, testFile.name+"_writed")}
	require.Nil(t, notExistingCF.checkWrite(), "Not existing filename")
	notExistingCF.Close()

	dstFile := structFile{path: wrongName}
	require.Error(t, dstFile.checkWrite(), "Testing create file in %s", wrongName)
	dstFile.Close()
	cleanupTestEnviromnent(workDir)
}

func TestCopy(t *testing.T) {
	//var test_file1, test_file2, wrong_file, test_dir string = LINUXTESTFILENAME, LINUXTESTFILENAME2, WIN_WRONG_FILENAME, LINUX_DIR
	workDir, err := createTestEnviromnent()
	if err != nil {
		t.Fatalf("Error on create test enviromnent")
	}
	wrongName := getWrongName(os_type)
	big_file, small_file := testFiles[1], testFiles[0]
	bigFileName := path.Join(workDir, big_file.name)

	require.Error(t, Copy(
		bigFileName+"_not_existed",
		wrongName,
		0, 100), "Open not existing file and try to write to forbidden dir")

	require.Nil(t, Copy(
		path.Join(workDir, small_file.name),
		path.Join(workDir, small_file.name+"_dst"),
		100, 0), "Copying created valid file to valid place")

	require.Error(t, Copy(
		bigFileName,
		wrongName,
		0, 10000), "Try to copy existing file to forbidden place")

	var offset, count uint64 = 0x100, 0
	require.Nil(t, Copy(
		bigFileName,
		bigFileName+"_dst",
		offset, count), "Test copying big file to valid place from offset")

	f, err := os.OpenFile(bigFileName+"_dst", os.O_RDONLY, 0666)
	if err != nil {
		t.Fatalf("Error open writed file for test it size")
	}
	writedFileState, err := f.Stat()
	if err != nil {
		t.Fatal("Error receive stat from writed file")
	}

	require.Equal(t, uint64(writedFileState.Size()), big_file.size-offset, "Testing then len(writed file) == len(source file)")
	f.Close()

	cleanupTestEnviromnent(workDir)

}
