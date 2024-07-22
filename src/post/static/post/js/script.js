function run() {
    const toggles = document.getElementsByClassName("tog");
    for ( toggle of toggles  ) {
        toggle.addEventListener("click", (e) => {
            hideOrShow(e)
        });
    }
}

function hideOrShow(event) {
    let element = event.target;
    let listId = "com_" + element.id;
    let selectedElements = document.getElementsByClassName(element.id);
    for ( child of selectedElements ) {
        child.classList.toggle("noshow");
    }
    const main = document.getElementById(listId);
    for ( const child of main.children ) {
        if (child.tagName == "ARTICLE" || child.tagName == "FOOTER") {
            child.classList.toggle("noshow");
        }
    }
    let numComments = main.getAttribute("n");
    if ( element.textContent.length == 3 ) {
        element.textContent = "[" + numComments + " more]";
    } else {
        element.textContent = "[-]";
    }
}

window.addEventListener("load", run);
