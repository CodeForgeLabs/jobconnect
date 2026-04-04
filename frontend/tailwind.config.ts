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
      },
      screens: {
        mobile: "0px",
        tablet: "580px",
        pc : "1100px",   
      },
      fontFamily: {
        sans: ['Inter', 'sans-serif'], 
        roboto: ['Roboto', 'sans-serif'],
      },
    },
  },
  plugins: [],
}

export default config