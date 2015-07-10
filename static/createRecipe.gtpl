{{ define "content" }}

    <div class="continer" style="width: 50%; margin: 0 auto;">
      <h2>ORBIT LABS</h2>
      <form action="/createrecipe/" method="POST">
      <table style="align: right">
        <tr>
          <td>Create Recipe:</br></td>
        </tr>
        <tr>
          <td>Recipe Name: </td>
          <td><input name="recipename" class="form-control" type="text" placeholder="Recipe Name"> </td>
        </tr>
        <tr>
          <td>Description: </td>
          <td><input name="description" class="form-control" type="text" placeholder="Description"> </td>
        </tr>
        <tr>
          <td>Variables: </td>
        </tr>
        <tr>
          <td>&nbsp</td><td><input style="text-align: right" name="var1" class="form-control" type="text" placeholder="Variable ID"> </td>
        </tr>
        <tr>
          <td>&nbsp</td><td><input style="text-align: right" name="var2" class="form-control" type="text" placeholder="Variable ID"> </td>
        </tr>
        <tr>
          <td>&nbsp</td><td><input style="text-align: right" name="var3" class="form-control" type="text" placeholder="Variable ID"> </td>
        </tr>
        <tr>
          <td>&nbsp</td><td><input style="text-align: right" name="var4" class="form-control" type="text" placeholder="Variable ID"> </td>
        </tr>
        <tr>
          <td>&nbsp</td><td><input style="text-align: right" name="var5" class="form-control" type="text" placeholder="Variable ID"> </td>
        </tr>
        <tr>
          <td><a href="/main/" >Back</a></td><td><button name="submit">Create</button></textarea></td>
        </tr>

      </table>
    </div>
  
{{ end }}