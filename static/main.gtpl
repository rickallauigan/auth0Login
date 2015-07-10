{{ define "content" }}

    <div class="container">
      <h2>ORBIT LABS</h2>
      <h3>Welcome {{.UserName}} !</h3>
      <table>
        <tr>
          <td><div><a href="/createrecipe/">Create Recipe</button></div></td>
        </tr>
        <tr>
          <td><div><a href="/apiv1/recipe/">View Recipe List</button></div></td>
        </tr>
        <tr>
          <td><div><a href="/master/">View Master List</button></div></td>
        </tr>
      </table>
    </div>
  
{{ end }}