import { useEffect, useState } from "react";
import { apiBaseUrl } from "../api";
import type { TransferJob } from "../types";

export function useSSE(jobId: string | null) {
  const [progress, setProgress] = useState<TransferJob | null>(null);
  const [done, setDone] = useState(false);

  useEffect(() => {
    if (!jobId) {
      return;
    }
    const es = new EventSource(`${apiBaseUrl()}/api/transfer/progress/${jobId}`, { withCredentials: true });
    es.onmessage = (e) => {
      const next = JSON.parse(e.data) as TransferJob;
      setProgress(next);
      if (next.status === "done" || next.status === "error") {
        setDone(true);
        es.close();
      }
    };
    es.onerror = () => {
      es.close();
    };
    return () => es.close();
  }, [jobId]);

  return { progress, log: progress?.log ?? [], done };
}

