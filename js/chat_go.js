$( document ).ready(function() {

	var conexion_final;

	$('#form_registro').on('submit', function(e) {
		e.preventDefault();
		user_name = $('#user_name').val()

		var conexion = new WebSocket("ws://localhost:8000/ws/" + user_name);    	
		conexion_final = conexion;
			conexion.onopen = function(){
				conexion.onmessage = function(response){
				console.log(response)
				val = $("#chat_area").val();
		   		$("#chat_area").val(val + "\n" + response.data); 
				}
			}
		$("#registro").hide();
   		$("#container_chat").show();
    });

    $('#form_message').on('submit', function(e) {
    	e.preventDefault();
    	conexion_final.send($('#msg').val());
    	$('#msg').val("")
    });
});