import { motion } from "framer-motion";
import type { TransferJob, TransferLog } from "../../types";

type ProgressViewProps = {
  progress: TransferJob | null;
  log: TransferLog[];
  done: boolean;
  onTransferMore: () => void;
};

export function ProgressView({
  progress,
  log,
  done,
  onTransferMore,
}: ProgressViewProps) {
  const progressCurrent = progress?.current ?? 0;
  const progressTotal = progress?.total ?? 0;
  const progressWidth = `${(progressCurrent / Math.max(progressTotal, 1)) * 100}%`;

  return (
    <main className="app-shell mx-auto min-h-screen max-w-5xl px-4 py-8 text-zinc-100">
      <h1 className="display-title text-4xl font-bold md:text-5xl">Transfer Progress</h1>
      <div className="mt-5 h-4 overflow-hidden rounded-full border border-zinc-600 bg-zinc-900/80 p-[2px]">
        <motion.div
          className="h-full rounded-full bg-gradient-to-r from-yandex to-soundcloud"
          animate={{ width: progressWidth }}
        />
      </div>
      <p className="mt-2 text-sm text-zinc-400">
        {progressCurrent}/{progressTotal} · transferred:{" "}
        {progress?.transferred ?? 0} · not found: {progress?.not_found ?? 0} ·
        {" "}errors: {progress?.errors ?? 0}
      </p>

      <div className="surface-panel mt-6 max-h-96 overflow-auto rounded-2xl p-4">
        {log.map((item, i) => (
          <div
            className="border-b border-zinc-800 py-2 text-sm"
            key={`${item.track_title}-${i}`}
          >
            <span className="mr-2">{item.status === "found" ? "?" : "?"}</span>
            {item.track_title}
          </div>
        ))}
      </div>

      {done && (
        <div className="mt-6 flex gap-3">
          {progress?.result_url && (
            <a
              className="btn btn-soundcloud px-4 py-2"
              href={progress.result_url}
              target="_blank"
              rel="noreferrer"
            >
              Open playlist in SoundCloud ?
            </a>
          )}
          <button
            className="btn btn-ghost px-4 py-2"
            onClick={onTransferMore}
          >
            Transfer more
          </button>
        </div>
      )}
    </main>
  );
}
