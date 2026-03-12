import { PlaylistGrid } from "../PlaylistGrid";
import type { Playlist } from "../../types";

type LibraryViewProps = {
  playlists: Playlist[];
  selected: Set<string>;
  onToggle: (id: string) => void;
  onSelectAll: () => void;
  onDeselectAll: () => void;
  preserveOriginalNames: boolean;
  onPreserveOriginalNamesChange: (next: boolean) => void;
  playlistName: string;
  onPlaylistNameChange: (next: string) => void;
  selectedCount: number;
  canTransfer: boolean;
  onStartTransfer: () => void;
  onBack: () => void;
};

export function LibraryView({
  playlists,
  selected,
  onToggle,
  onSelectAll,
  onDeselectAll,
  preserveOriginalNames,
  onPreserveOriginalNamesChange,
  playlistName,
  onPlaylistNameChange,
  selectedCount,
  canTransfer,
  onStartTransfer,
  onBack,
}: LibraryViewProps) {
  return (
    <main className="app-shell mx-auto min-h-screen max-w-6xl px-4 py-8 text-zinc-100">
      <div className="flex items-center justify-between gap-3">
        <h1 className="display-title text-4xl font-bold md:text-5xl">Выберите плейлисты</h1>
        <button
          className="btn btn-ghost text-sm"
          onClick={onBack}
        >
          Назад
        </button>
      </div>
      <div className="mt-3 flex gap-2">
        <button
          className="btn btn-ghost"
          onClick={onSelectAll}
        >
          Select All
        </button>
        <button
          className="btn btn-ghost"
          onClick={onDeselectAll}
        >
          Deselect All
        </button>
      </div>

      <div className="mt-6">
        <PlaylistGrid
          playlists={playlists}
          selected={selected}
          onToggle={onToggle}
        />
      </div>

      <div className="surface-panel mt-6 rounded-2xl p-5">
        <label
          htmlFor="preserve-original-names"
          className="flex items-center gap-2 text-sm text-zinc-300"
        >
          <input
            id="preserve-original-names"
            name="preserve_original_names"
            type="checkbox"
            checked={preserveOriginalNames}
            onChange={(e) => onPreserveOriginalNamesChange(e.target.checked)}
          />
          Сохранять исходные названия плейлистов
        </label>
        <label
          htmlFor="soundcloud-playlist-name"
          className="mt-4 block text-sm text-zinc-400"
        >
          Новое имя плейлиста в SoundCloud
        </label>
        <input
          id="soundcloud-playlist-name"
          name="soundcloud_playlist_name"
          className="input-field mt-2"
          value={playlistName}
          onChange={(e) => onPlaylistNameChange(e.target.value)}
          disabled={preserveOriginalNames}
        />
        <button
          className="btn mt-4 w-full border-0 bg-gradient-to-r from-yandex to-soundcloud py-3 text-black disabled:opacity-50"
          disabled={!canTransfer || selectedCount === 0}
          onClick={onStartTransfer}
        >
          Transfer {selectedCount} playlists {"->"}
        </button>
      </div>
    </main>
  );
}
