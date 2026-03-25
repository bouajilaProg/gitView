module.exports = {
  darkMode: 'class',
  content: ['./src/core/viewer/**/*.{html,js}'],
  theme: {
    extend: {
      colors: {
        'theme-bg': 'var(--bg)',
        'theme-panel': 'var(--panel)',
        'theme-border': 'var(--border)',
        'theme-text': 'var(--text)',
        'theme-muted': 'var(--muted)',
        'theme-dim': 'var(--dim)',
        'theme-blue': 'var(--blue)',
        'theme-green': 'var(--green)',
        'theme-yellow': 'var(--yellow)',
        'theme-red': 'var(--red)',
        'theme-magenta': 'var(--magenta)',
        'theme-cyan': 'var(--cyan)',
        'theme-orange': 'var(--orange)'
      }
    }
  }
}
