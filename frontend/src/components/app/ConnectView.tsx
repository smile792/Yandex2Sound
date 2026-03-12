import { ConnectCard, type ConnectCardProps } from "./ConnectCard";

type ConnectViewProps = ConnectCardProps & {
  canGoLibrary: boolean;
  showConnectScreen: boolean;
  onGoToLibrary: () => void;
};

export function ConnectView({
  canGoLibrary,
  showConnectScreen,
  onGoToLibrary,
  ...connectCardProps
}: ConnectViewProps) {
  return (
    <main className="app-shell min-h-screen px-4 py-8 text-zinc-100">
      <ConnectCard {...connectCardProps} />
      {canGoLibrary && showConnectScreen && (
        <div className="mx-auto mt-4 flex max-w-4xl justify-end">
          <button
            className="btn btn-ghost text-sm"
            onClick={onGoToLibrary}
          >
            К выбору плейлистов
          </button>
        </div>
      )}
    </main>
  );
}
