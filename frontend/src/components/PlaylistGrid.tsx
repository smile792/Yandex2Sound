import { motion } from "framer-motion";
import type { Playlist } from "../types";

type Props = {
  playlists: Playlist[];
  selected: Set<string>;
  onToggle: (id: string) => void;
};

export function PlaylistGrid({ playlists, selected, onToggle }: Props) {
  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {playlists.map((pl, idx) => {
        const active = selected.has(pl.id);
        return (
          <motion.button
            key={pl.id}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: idx * 0.03 }}
            onClick={() => onToggle(pl.id)}
            className={`rounded-xl border p-4 text-left transition ${
              active ? "border-yandex bg-zinc-900" : "border-zinc-700 bg-card"
            }`}
          >
            <div className="mb-2 flex items-center justify-between">
              <h3 className="font-bold">{pl.title}</h3>
              <input checked={active} type="checkbox" readOnly />
            </div>
            <p className="text-sm text-zinc-400">{pl.track_count} tracks</p>
          </motion.button>
        );
      })}
    </div>
  );
}

