export type Track = {
  id: string;
  title: string;
  artists: string;
  album: string;
  duration_ms: number;
  cover_url: string;
};

export type Playlist = {
  id: string;
  title: string;
  track_count: number;
  cover_url: string;
  special?: boolean;
};

export type TransferLog = {
  track_title: string;
  status: "found" | "not_found" | "error";
};

export type TransferJob = {
  id: string;
  status: "pending" | "running" | "done" | "error";
  total: number;
  current: number;
  transferred: number;
  not_found: number;
  errors: number;
  log: TransferLog[];
  result_url: string;
  last_track: string;
};

