import type { Config } from 'tailwindcss'

const config: Config = {
  content: [
    "./app/**/*.{js,ts,jsx,tsx}",
    "./components/**/*.{js,ts,jsx,tsx}"
  ],
  theme: {
    extend: {
      colors: {
        jobBlue: "#1152D4",
       
      "on-primary-fixed": "#04006d",
        "on-tertiary": "#ffffff",
        "error-container": "#ffdad6",
        "surface-dim": "#ccdbf3",
        "on-primary-fixed-variant": "#373a9b",
        "error": "#ba1a1a",
        "tertiary-fixed": "#c6e7ff",
        "surface-container-low": "#eff4ff",
        "surface-bright": "#f8f9ff",
        "surface": "#f8f9ff",
        "primary": "#15157d",
        "primary-container": "#2e3192",
        "on-secondary-container": "#57587f",
        "secondary-fixed-dim": "#c2c3f0",
        "tertiary": "#002c40",
        "secondary-fixed": "#e1e0ff",
        "secondary": "#5a5b82",
        "on-primary": "#ffffff",
        "on-error": "#ffffff",
        "surface-container-lowest": "#ffffff",
        "on-surface": "#0d1c2e",
        "on-secondary-fixed-variant": "#424369",
        "on-secondary-fixed": "#16183b",
        "on-secondary": "#ffffff",
        "secondary-container": "#d0d1fe",
        "surface-variant": "#d5e3fc",
        "on-surface-variant": "#464652",
        "inverse-surface": "#233144",
        "primary-fixed": "#e1e0ff",
        "on-tertiary-container": "#1fb4f7",
        "surface-container": "#e6eeff",
        "on-primary-container": "#9da1ff",
        "surface-container-highest": "#d5e3fc",
        "primary-fixed-dim": "#c0c1ff",
        "inverse-on-surface": "#eaf1ff",
        "on-tertiary-fixed-variant": "#004c6c",
        "on-error-container": "#93000a",
        "surface-container-high": "#dce9ff",
        "background": "#f8f9ff",
        "tertiary-fixed-dim": "#83cfff",
        "outline": "#777683",
        "on-tertiary-fixed": "#001e2e",
        "outline-variant": "#c7c5d4",
        "tertiary-container": "#00435f",
        "inverse-primary": "#c0c1ff",
        "surface-tint": "#4f54b4",
        "on-background": "#0d1c2e"
      },
      borderRadius: {
        "DEFAULT": "1rem",
        "lg": "2rem",
        "xl": "3rem",
        "full": "9999px"
      },
      screens: {
        mobile: "0px",
        tablet: "580px",
        pc : "1000px",   
      },
      fontFamily: {
        sans: ['Inter', 'sans-serif'], 
        roboto: ['Roboto', 'sans-serif'],
      },
      fontSize: {
        xs: '0.7rem',
        sm: '0.75rem',
        base: '0.875rem',
        lg: '1rem',
        xl: '1.125rem',
        '2xl': '1.25rem',
        '3xl': '1.5rem',
        '4xl': '1.875rem',
        '5xl': '2.25rem',
        '6xl': '3rem',
      },
    },
  },
  plugins: [],
}

export default config