package utils

func IsAllowedFileSize(fileSize, maxUploadSize int64) bool {
	return fileSize < maxUploadSize
}

func IsValidPhotoFormat(filetype string) bool {
	return filetype == "image/jpeg" || filetype == "image/jpg" || filetype == "image/png" || filetype == "image/gif"
}
