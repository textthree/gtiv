package obs

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/text3cn/goodle/providers/goodlog"
	"os"
)

// https://learn.microsoft.com/zh-cn/azure/storage/blobs/storage-quickstart-blobs-go?tabs=roles-azure-portal

type AzureClient struct{}

type azure struct {
	containerName string
	client        *azblob.Client
}

// endpoint = ctx.Config.Get("app.azure.cdn").ToString()
// connStr = ctx.Config.Get("app.azure.connStr").ToString()
// azure = &obs.AzureClient{}
// client := azure.Get(connStr, containName)
func (self *AzureClient) Get(connStr, containerName string) *azure {
	client, err := azblob.NewClientFromConnectionString(connStr, nil)
	if err != nil {
		goodlog.Error(err)
	}
	return &azure{containerName: containerName, client: client}
}

// contentType:
// m3u8: application/x-mpegurl
// png: image/png
func (self *azure) AzureUploadBuffer(blobName string, buffer []byte, contentType ...string) error {
	ctx := context.Background()
	options := azblob.UploadBufferOptions{}
	if len(contentType) > 0 {
		options.HTTPHeaders = &blob.HTTPHeaders{
			BlobContentType: to.Ptr(contentType[0]),
		}
	}
	_, err := self.client.UploadBuffer(ctx, self.containerName, blobName, buffer, &options)
	if err != nil {
		goodlog.Pink(err)
		return err
	}
	return nil
}

func (self *azure) AzureUploadFile(blobName string, file *os.File) error {
	ctx := context.Background()
	_, err := self.client.UploadFile(ctx, self.containerName, blobName, file, &azblob.UploadBufferOptions{})
	if err != nil {
		goodlog.Pink(err)
		return err
	}
	return nil
}
