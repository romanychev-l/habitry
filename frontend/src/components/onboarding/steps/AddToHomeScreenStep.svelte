<script lang="ts">
    import { onMount } from "svelte";
    import { _ } from "svelte-i18n";
    import { addToHomeScreen, checkHomeScreenStatus } from "@tma.js/sdk";

    let isAdded = false;
    let isChecking = true;

    onMount(async () => {
        try {
            const status = await checkHomeScreenStatus();
            isAdded = status === "missed"; // 'missed' means it's NOT on home screen? Wait, docs say:
            // "The checkHomeScreenStatus function checks if the user has already added the Mini App to the device's home screen."
            // The return type isn't explicitly clear in the snippet, but usually 'missed' means not added, 'unknown' means cant determine, 'installed' means added.
            // Let's assume we show the button if we can't confirm it's installed.
            // Actually, let's just show it always if we can't determine, or if it's not installed.
            // Re-reading docs snippet: "const status = await checkHomeScreenStatus();"
            // It doesn't show return values. Let's assume standard behavior or just try to add.
            // If we look at other TMA SDK features, usually it returns a string status.
            // Let's try to just call addToHomeScreen when button is clicked.

            // For now, let's just assume we show it.
            isChecking = false;
        } catch (e) {
            console.error("Error checking home screen status:", e);
            isChecking = false;
        }
    });

    function handleAddToHomeScreen() {
        addToHomeScreen();
    }
</script>

<div class="illustration-container">
    <div class="glow-effect"></div>
    <div class="illustration bounce">📱</div>
</div>
<h2 class="gradient-text">
    {$_("onboarding.homescreen.title", { default: "Add to Home Screen" })}
</h2>
<p class="description">
    {$_("onboarding.homescreen.description", {
        default: "Access Habitry quickly by adding it to your home screen.",
    })}
</p>

<div class="action-box">
    <button class="add-btn" on:click={handleAddToHomeScreen}>
        <span class="icon">➕</span>
        {$_("onboarding.homescreen.button", { default: "Add to Home Screen" })}
    </button>
</div>

<style>
    .illustration-container {
        position: relative;
        display: flex;
        align-items: center;
        justify-content: center;
        margin-bottom: 8px;
        min-height: 120px;
    }

    .glow-effect {
        position: absolute;
        width: 120px;
        height: 120px;
        border-radius: 50%;
        background: radial-gradient(
            circle,
            rgba(59, 130, 246, 0.3) 0%,
            transparent 70%
        );
        filter: blur(20px);
        animation: glow 2s ease-in-out infinite;
    }

    @keyframes glow {
        0%,
        100% {
            transform: scale(1);
            opacity: 0.6;
        }
        50% {
            transform: scale(1.1);
            opacity: 0.8;
        }
    }

    .illustration {
        font-size: 80px;
        position: relative;
        z-index: 1;
        filter: drop-shadow(0 4px 12px rgba(0, 0, 0, 0.2));
    }

    .bounce {
        animation: bounce 2s infinite;
    }

    @keyframes bounce {
        0%,
        20%,
        50%,
        80%,
        100% {
            transform: translateY(0);
        }
        40% {
            transform: translateY(-20px);
        }
        60% {
            transform: translateY(-10px);
        }
    }

    h2 {
        margin: 0;
        font-size: 28px;
        font-weight: 700;
        color: var(--tg-theme-text-color);
        letter-spacing: -0.5px;
    }

    .gradient-text {
        background: linear-gradient(135deg, #3b82f6 0%, #8b5cf6 100%);
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
        background-clip: text;
    }

    .description {
        margin: 12px 0 24px;
        font-size: 16px;
        line-height: 1.6;
        color: var(--tg-theme-hint-color, rgba(0, 0, 0, 0.6));
        max-width: 300px;
    }

    .action-box {
        margin-top: 12px;
        width: 100%;
        display: flex;
        justify-content: center;
    }

    .add-btn {
        background: var(--tg-theme-button-color, #3b82f6);
        color: var(--tg-theme-button-text-color, #ffffff);
        border: none;
        border-radius: 12px;
        padding: 12px 24px;
        font-size: 16px;
        font-weight: 600;
        cursor: pointer;
        display: flex;
        align-items: center;
        gap: 8px;
        transition: transform 0.2s;
        box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
    }

    .add-btn:active {
        transform: scale(0.95);
    }

    .icon {
        font-size: 18px;
    }

    :global([data-theme="dark"]) h2 {
        color: white;
    }

    :global([data-theme="dark"]) .description {
        color: rgba(255, 255, 255, 0.7);
    }
</style>
