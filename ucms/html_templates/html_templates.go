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
    <link href="https://cdnjs.cloudflare.com/ajax/libs/bootstrap/5.3.3/css/bootstrap.min.css" rel="stylesheet">
    <!-- Include MDBootstrap CSS -->
    <link href="https://cdnjs.cloudflare.com/ajax/libs/mdbootstrap/4.20.0/js/mdb.min.js" rel="stylesheet">
    <!-- Include FontAwesome CSS -->
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/5.9.0/css/all.min.css" rel="stylesheet">
  </head>
  <body>
    {{ .Body }}
    <script src="https://cdnjs.cloudflare.com/ajax/libs/bootstrap/5.3.3/js/bootstrap.min.js"></script>
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
  </head>                                                                                                                                                                                      <body>
  <h1>WYSIWYG Editor Example</h1>
  <form action="/submit" method="POST">
    <textarea id="editor" name="content"></textarea>
    <button type="submit">Submit</button>
  </form>
  <script>
  // Initialize CKEditor with source editing enabled
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
  </script>
  </body>
  </html>
`

const Submit = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>reCAPTCHA Form with Honeypot</title>
    <!-- Include reCAPTCHA script -->
    <script src="https://www.google.com/recaptcha/api.js?render={{ .Recaptchav3SiteKey }}"></script>
    <style>
        .honeypot {
            display: none; /* Hide the honeypot field */
        }
    </style>
</head>
<body>
    <h2>reCAPTCHA Form with Honeypot</h2>
    <form action="/submit" method="post">
        <label for="name">Name:</label><br>
        <input type="text" id="name" name="name"><br>
        <label for="email">Email:</label><br>
        <input type="email" id="email" name="email"><br><br>
        <!-- Add reCAPTCHA widget -->
        <input type="hidden" id="g-recaptcha-response" name="g-recaptcha-response">
        <button type="submit">Submit</button>
        <br>
        <!-- Add honeypot field -->
        <input type="text" class="honeypot" name="honeypot" autocomplete="off">
    </form>
    <script>
        // Execute when the form is submitted
        document.getElementById("myForm").addEventListener("submit", function(event) {
            event.preventDefault(); // Prevent the form from submitting normally

            // Execute reCAPTCHA verification
            grecaptcha.ready(function() {
                grecaptcha.execute('{{ .Recaptchav3SiteKey }}', {action: 'submit'}).then(function(token) {
                    // Set the token in the hidden input field
                    document.getElementById('g-recaptcha-response').value = token;
                    // Now submit the form
                    document.getElementById("myForm").submit();
                });
            });
        });
    </script>
</body>
</html>
`
const ySubmit = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>reCAPTCHA v3 Example</title>
    <!-- Include reCAPTCHA script -->
    <script src="https://www.google.com/recaptcha/api.js?render=6LdOMrgpAAAAAFzpFRQNsDqsLZdXTT3jIyPKReAD"></script>
</head>
<body>
    <h1>reCAPTCHA v3 Example</h1>
    <form id="myForm" action="/submit-form" method="POST">
        <!-- Add reCAPTCHA widget -->
        <input type="hidden" id="g-recaptcha-response" name="g-recaptcha-response">
        <button type="submit">Submit Form</button>
    </form>

    <script>
        // Execute when the form is submitted
        document.getElementById("myForm").addEventListener("submit", function(event) {
            event.preventDefault(); // Prevent the form from submitting normally

            // Execute reCAPTCHA verification
            grecaptcha.ready(function() {
                grecaptcha.execute('6LdOMrgpAAAAAFzpFRQNsDqsLZdXTT3jIyPKReAD', {action: 'submit'}).then(function(token) {
                    // Set the token in the hidden input field
                    document.getElementById('g-recaptcha-response').value = token;
                    // Now submit the form
                    document.getElementById("myForm").submit();
                });
            });
        });
    </script>
</body>
</html>
`
