let socket = new WebSocket("ws://127.0.0.1:7000/ws");

socket.onopen = function(e) {
  socket.send(JSON.stringify({"stage": "larissa-onboarding"}));
};

socket.onmessage = function(event) {
    msg = JSON.parse(event.data)
    if (msg.stage) {
      activeStage = msg
      presentStage(activeStage)
    }
    updateStats(msg)
};

socket.onclose = function(event) {
  if (~event.wasClean) {
    console.log('[close] Connection died');
  }
};

socket.onerror = function(error) {
  console.log(`[error] ${error.message}`);
};


function updateStats(stat) {
  const peopleDiv = document.getElementById("people")
  peopleDiv.innerHTML = ""
  n = 5
  if (stat.People.length < 5 ) {
      n = stat.People.length
  }
  for (let item=0; item < n; item++) {
      p = document.createElement("h2");
      setContentAndToken(p,stat.People[item])
      peopleDiv.appendChild(p);
  }

  const stagesDiv = document.getElementById("stages")
    stagesDiv.innerHTML = ""
    n = 5
    if (stat.Stages.length < 5 ) {
        n = stat.Stages.length
    }
    for (let item=0; item < n; item++) {
        p = document.createElement("h2");
        setContentAndToken(p,stat.Stages[item])
        stagesDiv.appendChild(p);
    }


}

let activeStage = {
  onboarding: "larissa",
  stage: {
    name: "larissa onboarding",
    token: "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
  },
  owner: {
    name: "aereum-org",
    token: "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
  },
  moderators: [
    {
      name: "ruben",
      token: "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
    },
    {
      name: "larissa",
      token: "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
    }
  ],
  submittors: [],
  content: [
    {
      author: {
        name: "aereum-org",
        token: "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
      },
      html: "<p> Lorem__ ipsum dolor sit amet consectetur adipisicing elit. Magnam, adipisci, necessitatibus praesentium delectus nam, eum alias rem incidunt sapiente sint repudiandae quo? Cupiditate ratione magni inventore consectetur maiores nostrum ex?</p><p> Lorem ipsum, dolor sit amet consectetur adipisicing elit. Sit, omnis?</p>",
      timestamp : "2022-08-03T10:03:22"
    },
    {
      author: {
        name: "larissa",
        token: "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
      },
      html: "<p> Lorem__ ipsum dolor sit amet consectetur adipisicing elit. Reiciendis nostrum maxime veritatis similique. Iusto facere temporibus neque. Debitis, iusto beatae.</p>",
      timestamp : "2022-08-03T10:03:22"
    }
  ]
}

function setContentAndToken(el, obj) {
  el.textContent = obj.name
  el.addEventListener('click', tokenClick(obj.token))
}

function setContentAndTokenOnId(id, obj) {
  const el = document.getElementById(id)
  setContentAndToken(el, obj)
}

function appendElementWithContentAndTokenOnId(id, tag, obj) {
  const el = document.createElement(tag)
  setContentAndToken(el, obj)
  const container = document.getElementById(id)
  container.innerHTML = ""
  container.appendChild(el)
}

function appendElementsWithContentAndTokenOnId(id, tag, obj) {
  const container = document.getElementById(id)
  container.innerHTML = ""
  for (ind of obj) {
    const el = document.createElement(tag)
    setContentAndToken(el, ind)
    container.appendChild(el)
  }
}


function presentStage(msg) {
  setContentAndTokenOnId("stage_title", msg.stage)

  appendElementWithContentAndTokenOnId("stage_owner", "span", msg.owner)
  appendElementsWithContentAndTokenOnId("stage_moderators", "h2", msg.moderators)
  document.getElementById("message_board").innerHTML = ""
  for (let post of msg.content) {
      createContent(post, post.author.name !== msg.onboarding)
  }
}

function createContent(post, other) {
  const postElement = document.createElement("div")
  postElement.setAttribute('class', other ? 'post' : 'own-post')
  const name = document.createElement("h2")
  name.setAttribute('class', other ? 'name' : 'self')
  name.textContent = post.author.name
  setContentAndToken(name, post.author)
  postElement.appendChild(name)
  const content = document.createElement("div")
  content.innerHTML = post.html
  postElement.appendChild(content)
  const boardElement = document.getElementById("message_board")
  boardElement.appendChild(postElement)
}

presentStage(activeStage)

function tokenClick(token) {
  return () => {
    console.log(token)
    socket.send(JSON.stringify({"token": token}));
  }
}