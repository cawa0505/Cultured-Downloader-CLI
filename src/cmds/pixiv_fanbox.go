package cmds

import (
	"github.com/spf13/cobra"
	"github.com/KJHJason/Cultured-Downloader-CLI/api"
	"github.com/KJHJason/Cultured-Downloader-CLI/api/pixivfanbox"
	"github.com/KJHJason/Cultured-Downloader-CLI/gdrive"
	"github.com/KJHJason/Cultured-Downloader-CLI/request"
	"github.com/KJHJason/Cultured-Downloader-CLI/utils"
)

var (
	fanboxCookieFile     string
	fanboxSession        string
	fanboxCreatorIds     []string
	fanboxPageNums       []string
	fanboxPostIds        []string
	fanboxDlThumbnails   bool
	fanboxDlImages       bool
	fanboxDlAttachments  bool
	fanboxDlGdrive       bool
	fanboxGdriveApiKey   string
	fanboxOverwriteFiles bool
	pixivFanboxCmd       = &cobra.Command{
		Use:   "pixiv_fanbox",
		Short: "Download from Pixiv Fanbox",
		Long:  "Supports downloading from Pixiv by artwork ID, illustrator ID, tag name, and more.",
		Run: func(cmd *cobra.Command, args []string) {
			request.CheckInternetConnection()

			pixivFanboxConfig := api.Config{
				OverwriteFiles: fanboxOverwriteFiles,
			}
			if fanboxGdriveApiKey != "" {
				pixivFanboxConfig.GDriveClient = gdrive.GetNewGDrive(
					fanboxGdriveApiKey, 
					utils.MAX_CONCURRENT_DOWNLOADS,
				)
			}

			pixivFanboxDl := pixivfanbox.PixivFanboxDl{
				CreatorIds:      fanboxCreatorIds,
				CreatorPageNums: fanboxPageNums,
				PostIds:         fanboxPostIds,
			}
			pixivFanboxDl.ValidateArgs()

			pixivFanboxDlOptions := pixivfanbox.PixivFanboxDlOptions{
				DlThumbnails:    fanboxDlThumbnails,
				DlImages:        fanboxDlImages,
				DlAttachments:   fanboxDlAttachments,
				DlGdrive:        fanboxDlGdrive && pixivFanboxConfig.GDriveClient != nil,
				SessionCookieId: fanboxSession,
			}
			if fanboxCookieFile != "" {
				cookies, err := utils.ParseNetscapeCookieFile(
					fanboxCookieFile, 
					fanboxSession, 
					utils.PIXIV_FANBOX,
				)
				if err != nil {
					utils.LogError(
						err,
						"",
						true,
					)
				}
				pixivFanboxDlOptions.SessionCookies = cookies
			}
			pixivFanboxDlOptions.ValidateArgs()

			pixivfanbox.PixivFanboxDownloadProcess(
				&pixivFanboxConfig,
				&pixivFanboxDl,
				&pixivFanboxDlOptions,
			)
		},
	}
)

func init() {
	mutlipleIdsMsg := getMultipleIdsMsg()
	pixivFanboxCmd.Flags().StringVar(
		&fanboxSession,
		"session",
		"",
		"Your FANBOXSESSID cookie value to use for the requests to Pixiv Fanbox.",
	)
	pixivFanboxCmd.Flags().StringSliceVar(
		&fanboxCreatorIds,
		"creator_id",
		[]string{},
		utils.CombineStringsWithNewline(
			[]string{
				"Pixiv Fanbox Creator ID(s) to download from.",
				mutlipleIdsMsg,
			},
		),
	)
	pixivFanboxCmd.Flags().StringSliceVar(
		&fanboxPageNums,
		"page_num",
		[]string{},
		utils.CombineStringsWithNewline(
			[]string{
				"Min and max page numbers to search for corresponding to the order of the supplied Pixiv Fanbox creator ID(s).",
				"Format: \"num\", \"minNum-maxNum\", or \"\" to download all pages",
				"Leave blank to download all pages from each creator.",
			},
		),
	)
	pixivFanboxCmd.Flags().StringSliceVar(
		&fanboxPostIds,
		"post_id",
		[]string{},
		utils.CombineStringsWithNewline(
			[]string{
				"Pixiv Fanbox post ID(s) to download.",
				mutlipleIdsMsg,
			},
		),
	)
	pixivFanboxCmd.Flags().BoolVar(
		&fanboxDlThumbnails,
		"dl_thumbnails",
		true,
		"Whether to download the thumbnail of a Pixiv Fanbox post.",
	)
	pixivFanboxCmd.Flags().BoolVar(
		&fanboxDlImages,
		"dl_images",
		true,
		"Whether to download the images of a Pixiv Fanbox post.",
	)
	pixivFanboxCmd.Flags().BoolVar(
		&fanboxDlAttachments,
		"dl_attachments",
		true,
		"Whether to download the attachments of a Pixiv Fanbox post.",
	)
	pixivFanboxCmd.Flags().BoolVar(
		&fanboxDlGdrive,
		"dl_gdrive",
		true,
		"Whether to download the Google Drive links of a Pixiv Fanbox post.",
	)
	pixivFanboxCmd.Flags().StringVar(
		&fanboxGdriveApiKey,
		"gdrive_api_key",
		"",
		utils.CombineStringsWithNewline(
			[]string{
				"Google Drive API key to use for downloading gdrive files.",
				"Guide: https://github.com/KJHJason/Cultured-Downloader/blob/main/doc/google_api_key_guide.md",
			},
		),
	)
}
