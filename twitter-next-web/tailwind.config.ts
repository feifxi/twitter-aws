// tailwind.config.ts
import type { Config } from "tailwindcss";

const config: Config = {
    content: [
        "./src/**/*.{js,jsx,ts,tsx}",
        "./app/**/*.{js,jsx,ts,tsx}",
        "./pages/**/*.{js,jsx,ts,tsx}",
        "./components/**/*.{js,jsx,ts,tsx}",
    ],
    darkMode: "class", // enable class-based dark mode
    theme: {
        extend: {
            colors: {
                primary: "var(--color-primary)",
                background: "var(--color-background)",
                foreground: "var(--color-foreground)",
                card: "var(--color-card)",
                cardForeground: "var(--color-card-foreground)",
                border: "var(--color-border)",
                // add more mappings as needed
            },
        },
    },
    plugins: [],
};

export default config;
