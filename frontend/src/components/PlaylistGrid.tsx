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
            className={`surface-card rounded-2xl p-4 text-left transition duration-200 ${
              active
                ? "border-yandex bg-zinc-900/95 shadow-[0_0_0_1px_rgba(255,211,71,0.28),0_12px_24px_rgba(0,0,0,0.34)]"
                : "hover:border-zinc-500 hover:translate-y-[-2px]"
            }`}
          >
            <div className="mb-2 flex items-center justify-between">
              <h3 className="font-bold tracking-wide">{pl.title}</h3>
              <input
                id={`playlist-select-${pl.id}`}
                name="playlist_select"
                checked={active}
                type="checkbox"
                readOnly
              />
            </div>
            <p className="text-sm text-zinc-400">{pl.track_count} tracks</p>
          </motion.button>
        );
      })}
    </div>
  );
}

