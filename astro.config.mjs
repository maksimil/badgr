import { defineConfig } from "astro/config";
import Windicss from "vite-plugin-windicss";

// https://astro.build/config
export default defineConfig({
  vite: {
    plugins: [Windicss()],
  },
});
