import axios from 'axios';

export interface FileItem {
  name: string;
  path: string;
  is_dir: boolean;
  size: number;
  mod_time: number;
  capabilities: number;
}

export interface ListResponse {
  items: FileItem[];
  effective_path: string;
}

export const CAP_BROWSE = 1;
export const CAP_STREAM = 2;
export const CAP_RENDER = 4;
export const CAP_EDIT = 8;

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
  // Remove leading slash if present to avoid double slashes after /raw/
  const cleanPath = path.startsWith('/') ? path.slice(1) : path;
  // Use replaceAll instead of split/map/join to avoid multiple array allocations
  const encodedPath = encodeURIComponent(cleanPath).replaceAll('%2F', '/');
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
  // Use replaceAll instead of split/map/join to avoid multiple array allocations
  const encodedPath = encodeURIComponent(cleanPath).replaceAll('%2F', '/');
  const { data } = await api.get<string>(`/raw/${encodedPath}`, { responseType: 'text' });
  return data;
};

export default api;
