{{ define "content" }}
  <body>

    <div class="container">
      <h2>Login</h2>
      <form role="form" method="POST" action="/main">
        <div class="form-group">
          Username: <input name="username" type="text" class="form-control" placeholder="UserName">
          <br/>
          Password: <input name="password" type="password" class="form-control" placeholder="Password">
        </div>
        <button type="submit" class="btn btn-default">Log In</button>
      </form>
      <div class="form-group">
      </div>
    </div>
    
  </body>

{{ end }}