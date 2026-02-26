import { API_URL } from '@/config';
import { request } from '@umijs/max';
import { message } from 'antd';
import file2md5 from 'file2md5';
import md5 from 'md5';
import { UploadRequestOption } from 'rc-upload/lib/interface';

export enum minioUploadType {
  file = '1',
  video = '2',
  picture = '3',
}
export async function minioUpload(
  file: File,
  fileType: minioUploadType,
  onProgress?: (progress: number) => void,
) {
  const data = new FormData();
  data.append('file', file);
  data.append('fileType', fileType);
  data.append('operationID', new Date().getTime() + '');
  return request<{
    data: any;
  }>('/third/minio_upload', {
    method: 'POST',
    data,
    headers: {
      token: localStorage.getItem('IMAdminToken')!,
    },
    baseURL: API_URL,
    onUploadProgress(progressEvent) {
      const complete = (progressEvent.loaded / progressEvent.total) * 100 || 0;
      if (onProgress) {
        onProgress(complete);
      }
    },
  });
}

export async function getCosAuthorization() {
  return request('/api/third/tencent_cloud_storage_credential', {
    method: 'POST',
    data: {},
    baseURL: API_URL,
  });
}

export async function upload(file: any, hash: any) {
  return request('/object/initiate_multipart_upload', {
    method: 'POST',
    data: {
      hash,
      size: file.size,
      partSize: file.size,
      maxParts: -1,
      cause: 'test-file',
      name: file.name,
      contentType: 'application/octet-stream',
    },
    baseURL: API_URL,
  });
}

export async function completeUpload(uploadID: string, hash: string, fileName: string) {
  return request('/object/complete_multipart_upload', {
    method: 'POST',
    data: {
      uploadID,
      parts: [hash],
      name: localStorage.getItem('IMAdminUserID') + '/' + fileName,
      contentType: 'application/octet-stream',
    },
    baseURL: API_URL,
  });
}

export async function handleFileUpload(data: UploadRequestOption) {
  try {
    const fileSize = (data.file as File).size;
    if (fileSize > 1024 * 1024 * 1) {
      message.error('图片尺寸过大');
      return;
    }

    const baseHash = await file2md5(data.file as File);
    const hash = [baseHash].join(',');

    const uploadResponse = await upload(data.file, md5(hash));

    if (!uploadResponse.data.url) {
      const uploadID = uploadResponse.data.upload.uploadID;
      const fileName = (data.file as File).name;
      const partURL = uploadResponse.data.upload.sign.parts[0].url;

      await fetch(partURL, {
        method: 'PUT',
        body: data.file,
      });

      const completeResponse = await completeUpload(uploadID, hash, fileName);
      return completeResponse.data.url;
    } else {
      return uploadResponse.data.url;
    }
  } catch (error) {
    console.log(error);
    message.error('上传失败！');
  }
}
