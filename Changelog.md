### February 2026

New features:
- Add `InsertHTMLBox` for rendering HTML content into a rectangular area on the PDF.
  - Supports tags: `<b>`, `<i>`, `<u>`, `<br>`, `<p>`, `<h1>`–`<h6>`, `<font>`, `<span style>`, `<img>`, `<ul>`, `<ol>`, `<li>`, `<hr>`, `<center>`, `<a>`, `<blockquote>`, `<sub>`, `<sup>`
  - Supports inline CSS: color, font-size, font-family, font-weight, font-style, text-decoration, text-align
  - Supports CSS color formats: `#RGB`, `#RRGGBB`, `rgb(r,g,b)`, named colors
  - Automatic text wrapping, alignment (left/center/right), and list rendering
- Add `HTMLBoxOption` struct for configuring HTML rendering (font families, default size, colors, line spacing)
- Add built-in HTML parser (`parseHTML`) with support for nested tags, attributes, self-closing tags, and HTML entities

### May 2016

Remove old function
- ```GoPdf.AddFont(family string, ifont IFont, zfontpath string)```.
- Remove all font map file.
