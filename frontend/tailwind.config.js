const defaultTheme = require('tailwindcss/defaultTheme')

/** @type {import('tailwindcss').Config} */
module.exports = {
  // mode: 'jit',
  plugins: [require('@tailwindcss/forms'), require('@tailwindcss/typography')],
  content: [],
  theme: {
    extend: {
      fontFamily: {
        sans: ['iA Writer', ...defaultTheme.fontFamily.sans],
        // sans: ['InterVariable', ...defaultTheme.fontFamily.sans],
        raleway: ['RalewayVariable', ...defaultTheme.fontFamily.sans],
        // mono: ['Menlo', 'Monaco', 'Consolas', 'Liberation Mono', 'Courier New', 'monospace'],
      },
      colors: {
        ral: {
          // RAL Classic Weiß- und Schwarztöne
          9001: '#e9e0d2', // Cremeweiß
          9002: '#d7d5cb', // Grauweiß
          9003: '#ecece7', // Signalweiß
          9004: '#2b2b2c', // Signalschwarz
          9005: '#0e0e10', // Tiefschwarz
          9006: '#a1a1a0', // Weißaluminium   9006 + 9007 Farben stammen aus dem Korrosionsschutzprogramm der ehemaligen Reichsbahn (Deutsche Bundes Eisenbahn).
          9007: '#878581', // Graualuminium
          9010: '#f1ece1', // Reinweiß
          9011: '#27292b', // Graphitschwarz
          9016: '#f1f0ea', // Verkehrsweiß
          9017: '#2a292a', // Verkehrsschwarz
          9018: '#c8cbc4', // Papyrusweiß
          9022: '#858583', // Perlhellgrau
          9023: '#797b7a', // Perldunkelgrau

          // Farbe der Kategorie Grautöne, Teil der Sammlung RAL Classic.
          7000: '#7a888e', // Fehgrau
          7001: '#8c969d', // Silbergrau
          7002: '#817863', // Olivgrau
          7003: '#7a7669', // Moosgrau
          7004: '#9b9b9b', // Signalgrau
          7005: '#6c6e6b', // Mausgrau
          7006: '#766a5e', // Beigegrau
          7008: '#745e3d', // Khakigrau
          7009: '#5d6058', // Grüngrau
          7010: '#585c56', // Zeltgrau
          7011: '#52595d', // Eisengrau
          7012: '#575d5e', // Basaltgrau
          7013: '#575044', // Braungrau
          7015: '#4f5358', // Schiefergrau
          7016: '#383e42', // Anthrazitgrau
          7021: '#2f3234', // Schwarzgrau
          7022: '#4c4a44', // Umbragrau
          7023: '#808076', // Betongrau
          7024: '#45494e', // Graphitgrau
          7026: '#374345', // Granitgrau
          7030: '#928e85', // Steingrau
          7031: '#5b686d', // Blaugrau
          7032: '#b5b0a1', // Kieselgrau
          7033: '#7f8274', // Zementgrau
          7034: '#92886f', // Gelbgrau
          7035: '#c5c7c4', // Lichtgrau
          7036: '#979392', // Platingrau
          7037: '#7a7b7a', // Staubgrau
          7038: '#b0b0a9', // Achatgrau
          7039: '#6b665e', // Quarzgrau
          7040: '#989ea1', // Fenstergrau
          7042: '#8e9291', // Verkehrsgrau A
          7043: '#4f5250', // Verkehrsgrau B
          7044: '#b7b3a8', // Seidengrau
          7045: '#8d9295', // Telegrau 1
          7046: '#7f868a', // Telegrau 2
          7047: '#c8c8c7', // Telegrau 4
          7048: '#817b73', // Perlmausgrau
        },
        // transitionProperty: {
        //   height: "max-height",
        // },
        // boxShadow: {
        //   top: "2px 1px 3px 2px rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06)",
        // },
        // maxHeight: {
        //   0: "0",
        //   "1/4": "25%",
        //   "1/2": "50%",
        //   "3/4": "75%",
        //   "4/5": "80%",
        //   "7/8": "87.5%",
        //   full: "100%",
        // },
        screens: {
          print: { raw: 'print' },
        },
      },
    },
    // fontFamily: {
    //   sans: ["Graphik", "sans-serif"],
    //   serif: ["Merriweather", "serif"],
    // },
    // screens: {
    //   xs: "375px",
    //   ...defaultTheme.screens,
    // },
  },
}

// const colors = require("tailwindcss/colors");
