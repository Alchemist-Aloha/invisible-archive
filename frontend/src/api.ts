import axios from 'axios';

export interface FileItem {
  name: string;
  path: string;
  is_dir: boolean;
  size: number;
  mod_time: number;
  capabilities: number;
}

export const CAP_BROWSE = 1;
export const CAP_STREAM = 2;
export const CAP_RENDER = 4;
export const CAP_EDIT = 8;

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE || '/api',
});

export const fetchList = async (path: string): Promise<FileItem[]> => {
  const { data } = await api.get<FileItem[]>('/ls', { params: { path } });
  return data;
};

export const searchFiles = async (q: string): Promise<FileItem[]> => {
  const { data } = await api.get<FileItem[]>('/search', { params: { q } });
  return data;
};

export const getRawUrl = (path: string) => {
  return `${api.defaults.baseURL}/raw/${path}`;
};

export const getThumbUrl = (path: string) => {
  return `${api.defaults.baseURL}/thumb?path=${encodeURIComponent(path)}`;
};

export default api;
