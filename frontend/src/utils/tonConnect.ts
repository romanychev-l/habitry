import { TonConnectUI } from "@tonconnect/ui";

export const tonConnectUI = new TonConnectUI({
  manifestUrl: "https://romanychev-l.github.io/habitry_public/tonconnect-manifest.json",
  buttonRootId: "ton-connect-button" // ID контейнера для кнопки
});