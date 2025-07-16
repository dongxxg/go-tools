package support

import (
	"mime/multipart"
)

type AttachMentRequest interface {
	GetLocation(name string) string
}

type DiskFile struct {
	Dir  string
	Name string
	//File *File
}

func (df *DiskFile) GetLocation(name string) string {
	return df.Dir
}

type MemoryFile struct {
}

type Saver interface {
	Save(reader *multipart.Part)
}

/*func (diskFile *DiskFile) Save(reader *multipart.Part) {
	var err error

	file, err := ioutil.TempFile(file.Dir)
	if err != nil {
		return err
	}
	size, err := io.Copy(file, reader)
	if cerr := file.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(file.Name())
		return err
	}
	diskFile.File = file
	return nil
}*/
