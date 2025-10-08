function setupMap(peerId, routesJSON) {
  if ("geolocation" in navigator) {
    navigator.geolocation.getCurrentPosition(
      (position) => {
        let lat = position.coords.latitude
        let long = position.coords.longitude

        let map = L.map('map', {
          minZoom: 13,
          maxZoom: 20,
        });
      
        if (window.location.hash) {
          let parts = window.location.hash.substring(1).split('/');
          map.setView([parts[2], parts[1]], parts[0]);
        } else {
          map.setView([lat, long], 13);
        }

        L.maplibreGL({
          style: 'https://tiles.openfreemap.org/styles/bright',
        }).addTo(map)

        // 1. Create a custom pane on your map
        map.createPane('tooltipPaneIntro');

        // 2. Set the z-index so it appears above default layers (e.g., above markers but below popups)
        map.getPane('tooltipPaneIntro').style.zIndex = 650;

        // 3. Enable pointer events on this pane so it can capture clicks
        map.getPane('tooltipPaneIntro').style.pointerEvents = 'auto';


        var tooltip = L.tooltip([lat, long], {content: `<center><h1>Welcome to Cyber Trespass!</h1></center>
          <div class="tooltip-content">
            <h2>Intro</h2>
            <ul>
              <li>Give trespassing access across your land and get access to a global network of trespassing routes</li>
              <li>Go further, go far by creating new routes beyond state road network</li>
              <li>Meet and socialize with people in the network as you trespass their properties</li>
              <li>Make travelling safer by forming trust and making properties accessible for walking, running and cycling</li>
            </ul>
            <h2>How it works</h2>
            <ul>
              <li>Create a route which trespasses your property</li>
              <li>When 3 people have used it and verified it's okay your route gets verified</li>
              <li>You get access to the global Cyber Trespass route network</li>
              <li>Fill in more routes as you travel and make the globe a trespassable network for humans</li>
            </ul>
          </div>`, direction: 'top', permanent: true, interactive: true, pane: 'tooltipPaneIntro',  // <-- important: assign to custom pane
          className: 'intro-tooltip'}).addTo(map);

          tooltip.on('click', function(e) {
            tooltip.remove();
            L.DomEvent.stopPropagation(e);
          });
        
        let counter = 0

        const Status = Object.freeze({
          Unverified: "unverified",
          Verified: "verified"
        });

        if (routesJSON.length > 0) {
          // Parse the main JSON string
          const parsedData = JSON.parse(routesJSON);

          var routes

          // Parse nested stringified JSON fields
          routes = parsedData.map(item => ({
            _id: item._id,
            coordinates: item.coordinates,
            created_by: item.created_by,
            verifiers: item.verifiers,
            status: item.status,
          }));

          var haveAccess

          for (const route of routes) {
            if (route.created_by == peerId && route.status == Status.Verified) {
              haveAccess = true
              break;
            }
          }

          for (const route of routes) {
            if (route.status == Status.Unverified) {
              var markerPickup = L.marker([route.coordinates[0][0].toString(), route.coordinates[0][1].toString()]).addTo(map);
              markerPickup.bindPopup("<p><center><b>Route ID: "+route._id+"</b></center></p>", { "closeButton": false});
              
              var button
              if (route.created_by != peerId && !route.verifiers.includes(peerId)) {
                button = `<div>
                <button onclick="verifyRoute('${route._id}')">Verify</button>
                </div>`
              } else {
                button = ""
              }
            
              var markerDropoff = L.marker([route.coordinates[1][0].toString(), route.coordinates[1][1].toString()]).addTo(map);
              markerDropoff.bindPopup(`<center><b>Route ID: `+route._id+`
                  <p>Status: `+Status.Unverified+`</p>
                `+button+`
                </center></b>`, { "closeButton": false, "className": "popup"});
              L.polyline(route.coordinates, {color: 'blue'}).addTo(map);
            } else {
              if (haveAccess) {
                var markerPickup = L.marker([route.coordinates[0][0].toString(), route.coordinates[0][1].toString()]).addTo(map);
                markerPickup.bindPopup("<p><center><b>Route ID: "+route._id+"</b></center></p>", { "closeButton": false});
              
                var markerDropoff = L.marker([route.coordinates[1][0].toString(), route.coordinates[1][1].toString()]).addTo(map);
                markerDropoff.bindPopup(`<center><b>Route ID: `+route._id+`
                  <p>Status: `+Status.Verified+`</p>
                  </center></b>`, { "closeButton": false, "className": "popup"});
                L.polyline(route.coordinates, {color: 'green'}).addTo(map);
              }
            }
          }
        }

        var pickupLat, pickupLng, dropoffLat, dropoffLng

        function onMapClick(e) {
          switch (counter) {
            case 0:
              var marker = L.marker([e.latlng.lat.toString(), e.latlng.lng.toString()]).addTo(map);
              
              marker.on('click', function(e) {
                marker.remove();
                polyline.remove()
              });

              marker.bindPopup(`<p><center><b>
                  Pickup
                </b></center></p>
                <i>Wrong click? Click the marker again to remove it.</i>`, { "closeButton": false}).openPopup();
              pickupLat = e.latlng.lat.toString()
              pickupLng = e.latlng.lng.toString()
              counter++
              return
            case 1:
              var marker = L.marker([e.latlng.lat.toString(), e.latlng.lng.toString()]).addTo(map);

              marker.on('click', function(e) {
                marker.remove();
                polyline.remove()
              });

              dropoffLat = e.latlng.lat.toString()
              dropoffLng = e.latlng.lng.toString()
              var latlngs = [
                [pickupLat, pickupLng],
                [dropoffLat, dropoffLng]
              ]

              marker.bindPopup(`<p><center><b>Dropoff
                <br><br>
                `+createButton(latlngs)+`
                </p></center></b>
                <i>Wrong click? Click the marker again to remove it.</i>`, { "closeButton": false, "className": "popup"}).openPopup();
                
              var polyline = L.polyline(latlngs, {color: 'red'}).addTo(map);

              pickupLat = 0
              pickupLng = 0
              dropoffLat = 0
              dropoffLng = 0

              counter++
              return
            default: 
              counter = 0
              var marker = L.marker([e.latlng.lat.toString(), e.latlng.lng.toString()]).addTo(map);
              
              marker.on('click', function(e) {
                marker.remove();
                polyline.remove()
              });

              marker.bindPopup(`<p><center><b>
                  Pickup
                </b></center></p>
                <i>Wrong click? Click the marker again to remove it.</i>`, { "closeButton": false}).openPopup();
              pickupLat = e.latlng.lat.toString()
              pickupLng = e.latlng.lng.toString()
              counter++
              return
          }
        }

        map.on('click', onMapClick);
      },
      (error) => {
        alert("In order to use Cyber Trespasser location needs to be enabled");
      }
    );
  } else {
    console.error("Geolocation is not supported by this browser.");
  }
}

function createButton(latlngs) {
  // Serialize latlngs as JSON string to pass in inline handler
  const latlngsStr = JSON.stringify(latlngs);
  return `<button onclick='createRoute(${latlngsStr})'>Create Route</button>`;
}

function removeMarker(marker) {
  marker.remove();
}

if (window.location.hash === "#close") {
  window.location.replace("/");
}

function createRoute(latlngs) {
  const routeCreated = new CustomEvent('route-created', {
    detail: {
        latlngs: JSON.stringify(latlngs),
    }
  });

  var elementMap = document.getElementById("map")

  elementMap.dispatchEvent(routeCreated);
}

function verifyRoute(id) {
  const routeVerified = new CustomEvent('route-verified', {
    detail: {
        id: JSON.stringify(id),
    }
  });

  var elementMap = document.getElementById("map")

  elementMap.dispatchEvent(routeVerified);
}