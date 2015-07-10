{{ define "content" }}

    <div class="container">
      <h2>Master List</h2>

        <table>
            <tr>
              <td><pre><b>Variable Name</b></pre></td>
              <td><pre><b>Unit</b></pre></td>
              <td><pre><b>Description</b></pre></td>
              <td><pre><b>Validation</b></pre></td>
            </tr>
          {{range .}}
            <tr>
              <td><pre>{{.VariableName}}</pre></td>
              <td><pre>{{.Unit}}</pre></td>
              <td><pre>{{.Description}}</pre></td>
              <td><pre>{{.Validation}}</pre></td>
            </tr>
          {{end}}

      <form role="form" method="POST" action="/master/">
  			    <tr>
            <td><pre><b>Add Variable :</b></pre></td><td>&nbsp</td><td>&nbsp</td><td>&nbsp</td>
            </tr>
  		  		<tr>
  		  			<td><input name="variableName" type="text" class="form-control" placeholder="Variable Name"> </td>
  		  			<td><input name="variableUnit" type="text" class="form-control" placeholder="Unit"> </td>
              <td><input name="variableDescription" type="text" class="form-control" placeholder="Description"> </td>
  		  			<td><input name="variableValidation" type="text" class="form-control" placeholder="Validation"> </td>
  		  		</tr>
            <tr>
              <td><a href="/main">HOME</a></td><td>&nbsp</td><td>&nbsp</td>
              <td><button type="submit" class="btn btn-default">ADD</button></td>
            </tr>
  		  </table>
		  </form>
    </div>
	
{{ end }}