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
