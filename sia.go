package main

import "fmt"
import (
	ds "github.com/ipfs/go-datastore"
	modules "gitlab.com/NebulousLabs/Sia/modules"
	//api "gitlab.com/NebulousLabs/Sia/node/api"
	"bytes"
	//dsq "github.com/ipfs/go-datastore/query"
	client "gitlab.com/NebulousLabs/Sia/node/api/client"
	"os"
	"unicode/utf16"
	"unicode/utf8"
)

const (
	// listMax is the largest amount of objects you can request from S3 in a list
	// call.
	listMax = 1000

	// deleteMax is the largest amount of objects you can delete from S3 in a
	// delete objects call.
	deleteMax = 1000

	defaultWorkers = 100
)

func main() {
	//Test function

	//fmt.Printf(string(bytes[:]))
	//properBytes, _ := DecodeUTF16(bytes)

	//fmt.Printf(properBytes)
	//fmt.Printf(fmt.Sprintf("%s", bytes))

	//test(ds.RandomKey())
	siaStore := NewSiaStore(Config{"", "", "", 0, 0})
	key := ds.RandomKey()

	siaStore.Put(key, []byte("ABC€"))
	fmt.Println(key.String())

	bytes, err := siaStore.Get(key)
	if err == nil {
		fmt.Println(err.Error())
	} else {
		str, _ := DecodeUTF16(bytes)
		fmt.Println("data: " + str)
	}

}
func DecodeUTF16(b []byte) (string, error) {

	if len(b)%2 != 0 {
		return "", fmt.Errorf("Must have even length byte slice")
	}

	u16s := make([]uint16, 1)

	ret := &bytes.Buffer{}

	b8buf := make([]byte, 4)

	lb := len(b)
	for i := 0; i < lb; i += 2 {
		u16s[0] = uint16(b[i]) + (uint16(b[i+1]) << 8)
		r := utf16.Decode(u16s)
		n := utf8.EncodeRune(b8buf, r[0])
		ret.Write(b8buf[:n])
	}

	return ret.String(), nil
}
func test(k ds.Key) {
	fmt.Printf(k.String())
}

//Memory cache piece of data.
//Gets deletes from memory after 25 seconds or until garbage collection.
//This is meant only speeding up get/put operations.
//CacheEntry should contain all the K/V pairs.
//CacheEntry is the ENTIRE 4MB sia file block. these may contain many small K/V pairs.
//Must be in some sort of accessible format on the caching layer. Slow sync over to sia.
type CacheEntry struct {
	timestamp int
	payload   []byte
	key       ds.Key
}

type Config struct {
	Address     string
	APIPassword string
	Bucket      string //Context we are using
	Packing     int    //How many values should be packed into a block/experimental
	Workers     int
}

type SiaStore struct {
	Config
	client client.Client
}

//NewSiaStore Config
func NewSiaStore(conf Config) *SiaStore {

	if conf.Address == "" {
		conf.Address = "localhost:9980"
	}
	if conf.APIPassword == "" {
		conf.APIPassword = os.Getenv("SIA_API_PASSWORD")
	}

	clientInt := client.Client{
		Address:   conf.Address,
		Password:  conf.APIPassword,
		UserAgent: "Sia-Agent",
	}

	return &SiaStore{
		conf,
		clientInt,
	}
}

func (s *SiaStore) Get(k ds.Key) ([]byte, error) {
	path, _ := modules.NewSiaPath(k.String())
	bytes, err := s.client.RenterStreamGet(path)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (s *SiaStore) Put(k ds.Key, value []byte) error {
	reader := bytes.NewReader(value)
	path, _ := modules.NewSiaPath(k.String())
	err := s.client.RenterUploadStreamPost(reader, path, 10, 12, true)
	if err == nil {
		return err
	}

	/*_, err := s.client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(s.s3Path(k.String())),
		Body:   bytes.NewReader(value),
	})
	return parseError(err)*/
	return nil
}

func (s *SiaStore) Has(k ds.Key) (exists bool, err error) {
	_, err = s.GetSize(k)
	if err != nil {
		return false, err
	}
	return true, nil
}
func (s *SiaStore) GetSize(k ds.Key) (size int, err error) {
	path, _ := modules.NewSiaPath(k.String())
	rf, err := s.client.RenterFileGet(path)
	sizeint := rf.File.Filesize
	if err != nil {
		return 0, err
	}
	return int(sizeint), nil
}

func (s *SiaStore) Delete(k ds.Key) error {
	path, _ := modules.NewSiaPath(k.String())
	err := s.client.RenterDeletePost(path)
	return err
}

/*func (s *SiaStore) Query(q dsq.Query) (dsq.Results, error) {
	if q.Orders != nil || q.Filters != nil {
		return nil, fmt.Errorf("siads: filters or orders are not supported")
	}
	limit := q.Limit + q.Offset
	if limit == 0 || limit > listMax {
		limit = listMax
	}


}*/

/*func NewSiaDatastore(conf Config) (*SiaStore, error) {

}*/
