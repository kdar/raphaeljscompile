raphaeljscompile
----------------

This is an attempt to make a backend compiler for RaphaelJS.
Meaning, this will take some javascript code of generating non-animating SVG images, and return the SVG code.

Another idea is to use this project: https://github.com/sourcegraph/webloop. It's able to generate a static page like PhantomJS. On that note you could just use PhantomJS to compile the RaphaelJS.

### Install

If you don't have v8.go installed, this is not go-getable. Install v8.go first: [https://github.com/idada/v8.go](https://github.com/idada/v8.go)

### Usage

`raphaeljscompile --help`

### Example

    echo 'var paper = Raphael(10, 50, 320, 100);
     var circle = paper.circle(50, 40, 10);
     circle.attr("fill", "#f00");
     circle.attr("stroke", "#fff");' | ./raphaeljscompile 

Output:

    <svg height="100" version="1.1" width="320" xmlns="http://www.w3.org/2000/svg" style="-webkit-tap-highlight-color:rgba(0,0,0,0);overflow:hidden;position:absolute;left:10px;top:50px"><desc style="-webkit-tap-highlight-color:rgba(0,0,0,0)">Created with Raphaël 2.1.2</desc><defs style="-webkit-tap-highlight-color:rgba(0,0,0,0)"></defs><circle cx="50" cy="40" r="10" fill="#ff0000" stroke="#ffffff" style="-webkit-tap-highlight-color:rgba(0,0,0,0)"></circle></svg>


### Why?

I had a need to generate SVG images on the backend because of a technical issue I was having where an HTML to PDF converter could not run RaphaelJS. So my idea was to generate the SVGs in the backend and feed that to the HTML. I ended using HTML5 canvas instead to solve my problem, but I decided to finish this as a proof-of-concept.

### Notes

 - This is not well tested. 
 - Pull requests welcome.
 - I initially started implementing a simulated DOM, but it became way too compilcated. I decided to make it easier/hackish instead.
