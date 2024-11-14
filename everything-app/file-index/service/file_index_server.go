package service

import (
	"context"
	"crypto/md5"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"training/file-index/pb"

	"github.com/djherbis/times"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type FileDiscoveryServer struct {
	pb.UnimplementedFileIndexServer
	fileStore FileStore
}
type FileInfoMap map[string]*pb.FileAttr

var fileInfoCache FileInfoMap

func NewFileDiscoveryServer(fileStore FileStore) *FileDiscoveryServer {
	return &FileDiscoveryServer{
		fileStore: fileStore,
	}
}

func (server *FileDiscoveryServer) GetCheckSumFiles(ctx context.Context, req *pb.CreateFileChecksumRequest) (*pb.CreateFileChecksumResponse, error) {
	filepaths := req.GetFilepath()
	var res = &pb.CreateFileChecksumResponse{Checksums: make(map[string]string)}

	for _, filepath := range filepaths {
		_, err := os.Stat(filepath)
		if err != nil {
			if os.IsPermission(err) {
				log.Printf("Permission denied")
				continue
			}
		}
		filename := strings.Split(filepath, "/")

		fileChecksum, err := calculateHash(filepath, md5.New)
		if err != nil {
			continue
		}
		log.Printf("receive a filepath checksum request with path :%s with checksum %s\n", filename[len(filename)-1], fileChecksum)
		res.Checksums[filepath] = fileChecksum
	}
	return res, nil
}

func (server *FileDiscoveryServer) ListFiles(req *pb.CreateFileDiscoverRequest, stream grpc.ServerStreamingServer[pb.CreateFileDiscoverResponse]) error {
	request := req.GetRequest()
	fmt.Printf("Receive request to list all files in computer %s\n", request)

	for {
		drives := getAvailableDrives()
		for _, drive := range drives {
			currentFiles := make(FileInfoMap)
			err := filepath.Walk(drive, func(path string, info fs.FileInfo, err error) error {
				if err != nil {
					if os.IsPermission(err) {
						fmt.Printf("Permission denied: %s\n", path)
						return nil
					}
					return err
				}
				// Skip directories
				if info.IsDir() {
					return nil
				}
				ext := filepath.Ext(path)

				// Get file timestamps
				createdAt, modifiedAt, accessedAt, err := getFileTimes(path)

				var content string
				switch ext {
				case ".docx":
					content, err = readDocxContent(path)
				case ".xlsx":
					content, err = readExcelContent(path)
				case ".pptx":
					content, err = readPPTXContent(path)
				}
				if err != nil {
					fmt.Printf("Failed to read %s file: %s, error: %v\n", ext, path, err)
					return nil
				}
				var attribute string
				flagAttributes, _ := IsHiddenFile(path)
				if flagAttributes {
					attribute = "Hidden"
				} else {
					attribute = "Read Only"
				}
				fileAttr := &pb.FileAttr{
					Path:       path,
					Name:       info.Name(),
					Type:       ext,
					Size:       info.Size(),
					CreatedAt:  timestamppb.New(createdAt),
					ModifiedAt: timestamppb.New(modifiedAt),
					AccessedAt: timestamppb.New(accessedAt),
					Attributes: attribute,
					Content:    content,
				}
				currentFiles[path] = fileAttr
				// Save file to store if not exists
				if _, exists := fileInfoCache[path]; !exists {
					if err := server.fileStore.Save(fileAttr); err != nil {
						return err
					}
				}

				res := &pb.CreateFileDiscoverResponse{
					Files: fileAttr,
				}
				if err := stream.Send(res); err != nil {
					return err
				}

				return nil
			})

			if err != nil {
				return err
			}

			// Check file deleted
			for path, fileAttr := range fileInfoCache {
				if _, exists := currentFiles[path]; !exists {
					if err := server.fileStore.Delete(fileAttr.Name); err != nil {
						return err
					}
				}
			}
			// Update cache
			fileInfoCache = currentFiles
		}
		time.Sleep(5 * time.Second)
	}
}

func getFileTimes(path string) (createdAt, modifiedAt, accessedAt time.Time, err error) {
	t, err := times.Stat(path)
	if err != nil {
		if os.IsPermission(err) {
			fmt.Printf("Permission denied: %s\n", path)
			return createdAt, modifiedAt, accessedAt, nil
		}
		return createdAt, modifiedAt, accessedAt, err
	}
	if t.HasBirthTime() {
		createdAt = t.BirthTime()
	}
	if t.HasChangeTime() {
		modifiedAt = t.ModTime()
	}
	accessedAt = t.AccessTime()

	// Return times with no errors
	return createdAt, modifiedAt, accessedAt, nil
}
