$(document).ready(function() {
    //var lock = new Auth0Lock(AUTH0_CLIENT_ID, AUTH0_DOMAIN);
	var lock = new Auth0Lock("mhqhf8fTNZKtDDZRdukygwWTbybVVHbC", "bukidutility.auth0.com");
	
    $('.btn-login').click(function(e) {
      e.preventDefault();
      lock.show({
        //callbackURL: "http://localhost:8080/oauth2callback"
        callbackURL: "http://bukidutility.appspot.com/oauth2callback"
      });
    });
});
