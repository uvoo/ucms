package html_templates 

const Markdown = `
  <!DOCTYPE html>
  <html lang="en">
  <head>
	  <meta charset="UTF-8">
	  <meta name="viewport" content="width=device-width, initial-scale=1.0">
	  <title>{{.Title}}</title>
  </head>
  <body>
	 {{ .Body }}
  </body>
  </html>
`

const MDBootstrap = `
  <!DOCTYPE html>
  <html lang="en">
  <head>
	  <meta charset="UTF-8">
	  <meta name="viewport" content="width=device-width, initial-scale=1.0">
	  <title>Hello World</title>
	  <!-- Include Bootstrap 5 CSS -->
	  <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0-alpha1/dist/css/bootstrap.min.css" rel="stylesheet">
	  <!-- Include MDBootstrap CSS -->
	  <link href="https://cdnjs.cloudflare.com/ajax/libs/mdbootstrap/4.19.1/css/mdb.min.css" rel="stylesheet">
	  <!-- Include FontAwesome CSS -->
	  <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/5.15.4/css/all.min.css" rel="stylesheet">
  </head>
  <body>
	 {{ .Body }}
	 <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0-alpha1/dist/js/bootstrap.bundle.min.js"></script>
  </body>
  </html>
`

const Error = `
  E: Unsupported template
`

const Editor = `
  <!DOCTYPE html>
  <html lang="en">
  <head>
	  <meta charset="UTF-8">
	  <meta name="viewport" content="width=device-width, initial-scale=1.0">
	  <title>WYSIWYG Editor Example</title>
	  <!-- Include CKEditor -->
	  <script src="https://cdn.ckeditor.com/ckeditor5/37.0.0/classic/ckeditor.js"></script>
  </head>                                                                                                                                                                                      <body>                                                                                                                                                                                           <h1>WYSIWYG Editor Example</h1>                                                                                                                                                              <form action="/submit" method="POST">
		  <textarea id="editor" name="content"></textarea>
		  <button type="submit">Submit</button>
	  </form>                                                                                                                                                                                                                                                                                                                                                                                   <script>                                                                                                                                                                                         // Initialize CKEditor with source editing enabled
		  ClassicEditor
			  .create(document.querySelector('#editor'), {
				  toolbar: ['', 'heading', '|', 'bold', 'italic', 'link', 'bulletedList', 'numberedList', '|', 'indent', 'outdent', '|', 'blockQuote', 'insertTable', '|', 'undo', 'redo', '|', 'source', 'uploadImage', 'blockQuote', 'codeBlock'],
  /*
  toolbar: {
	  items: [
		  'undo', 'redo',
		  '|', 'heading',
		  '|', 'fontfamily', 'fontsize', 'fontColor', 'fontBackgroundColor',
		  '|', 'bold', 'italic', 'strikethrough', 'subscript', 'superscript', 'code',
		  '|', 'link', 'uploadImage', 'blockQuote', 'codeBlock',
		  '|', 'bulletedList', 'numberedList', 'todoList', 'outdent', 'indent'
	  ],
	  shouldNotGroupWhenFull: false
  }
  */
				  language: 'en'
			  })
			  .catch(error => {
				  console.error(error);
			  });
	  </script>                                                                                                                                                                                </body>
  </html>
`