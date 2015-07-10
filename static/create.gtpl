{{ define "content" }}

    <div class="container">
      <h2>ORBIT LABS</h2>
    {{range .}}
      {{with .FarmID}}
        <p><b>{{.}}</b> wrote:</p>
      {{else}}
        <p>An anonymous person wrote:</p>
      {{end}}
      	<pre>{{.Token}}</pre>
		<pre>{{.TimeCreated}}</pre>
		<pre>{{.TimeStored}}</pre>
		<pre>{{.DeviceUse}}</pre>
		<pre>{{.VariableID}}</pre>
		<pre>{{.FarmID}}</pre>
		<pre>{{.Value}}</pre>
    {{end}}
    
    </div>
  
{{ end }}