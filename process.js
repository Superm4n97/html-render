const fs = require('fs');
const postcss = require('postcss');
const purgecss = require('@fullhuman/postcss-purgecss');

const cssFilePath = '/Users/mohin/go/src/github.com/html-render/pkg/template/static/assets/styles/main.scss'; // Path to the CSS file

// Read the CSS file
fs.readFile(cssFilePath, 'utf8', (err, css) => {
  if (err) {
    console.error(`Error reading the CSS file at ${cssFilePath}:`, err);
    return;
  }

  // Define PurgeCSS options
  const purgecssOptions = {
    content: ['./pkg/**/*.gohtml'], // Update this to the path where your Go HTML templates are located
    defaultExtractor: content => {
      // This extractor assumes your Go HTML templates use standard HTML structure
      // Modify if your templates contain Go-specific syntax that affects class or ID extraction
      return content.match(/[\w-/:]+(?<!:)/g) || [];
    },
    safelist: {
      standard: [
        'active',
        'is-visible',
        'is-right-0',
        'fserv-field',
        'select2-container',
        'select2',
        'fs-webform-container',
        'placeholder',
        'fserv-button-submit'
      ],
      deep: [
        /^fserv-/
      ],
      greedy: [
        // Add any greedy patterns if needed
      ],
      keyframes: true,
    }
  };

  // Process CSS with PostCSS and PurgeCSS
  postcss([purgecss(purgecssOptions)])
    .process(css, { from: cssFilePath })
    .then(result => {
      // Write the processed CSS back to the file
      fs.writeFile(cssFilePath, result.css, err => {
        if (err) {
          console.error(`Error writing the processed CSS back to ${cssFilePath}:`, err);
          return;
        }
        console.log('CSS processed successfully!');
      });
    })
    .catch(error => {
      console.error('Error during CSS processing with PostCSS and PurgeCSS:', error);
    });
});
