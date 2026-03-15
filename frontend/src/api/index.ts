import axios from 'axios';
import type { FileItem, ListResponse } from '../types';

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE || '/api',
});

export const fetchList = async (path: string): Promise<ListResponse> => {
  const { data } = await api.get<ListResponse>('/ls', { params: { path } });
  return data;
};

export const searchFiles = async (q: string): Promise<FileItem[]> => {
  const { data } = await api.get<FileItem[]>('/search', { params: { q } });
  return data;
};

export const getRawUrl = (path: string, download: boolean = false) => {
  const cleanPath = path.startsWith('/') ? path.slice(1) : path;
  const encodedPath = cleanPath.split('/').map(encodeURIComponent).join('/');
  let url = `${api.defaults.baseURL}/raw/${encodedPath}`;
  if (download) {
    url += '?download=1';
  }
  return url;
};

export const getThumbUrl = (path: string) => {
  return `${api.defaults.baseURL}/thumb?path=${encodeURIComponent(path)}`;
};

export const fetchText = async (path: string): Promise<string> => {
  const cleanPath = path.startsWith('/') ? path.slice(1) : path;
  const encodedPath = cleanPath.split('/').map(encodeURIComponent).join('/');
  const { data } = await api.get<string>(`/raw/${encodedPath}`, { responseType: 'text' });
  return data;
};

export default api;
