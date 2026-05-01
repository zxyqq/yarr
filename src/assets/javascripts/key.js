var helperFunctions = {
  scrollContent: function(direction) {
    var padding = 40
    var scroll = document.querySelector('.content')
    if (!scroll) return

    var height = scroll.getBoundingClientRect().height
    var newpos = scroll.scrollTop + (height - padding) * direction

    if (typeof scroll.scrollTo == 'function') {
      scroll.scrollTo({top: newpos, left: 0, behavior: 'smooth'})
    } else {
      scroll.scrollTop = newpos
    }
  }
}
var shortcutFunctions = {
  openItemLink: function() {
    if (vm.itemSelectedDetails && vm.itemSelectedDetails.link) {
      window.open(vm.itemSelectedDetails.link, '_blank', 'noopener,noreferrer')
    }
  },
  toggleReadability: function() {
    vm.toggleReadability()
  },
  toggleItemRead: function() {
    if (vm.itemSelected != null) {
      vm.toggleItemRead(vm.itemSelectedDetails)
    }
  },
  openShortcuts: function() {
    vm.showSettings('shortcuts')
  },
  closeModal: function() {
    if (vm.settings) {
      vm.settings = ''
      return
    }
    if (vm.itemSelected != null) {
      vm.itemSelected = null
    }
  },
  markAllRead: function() {
    // same condition as 'Mark all read button'
    if (vm.filterSelected == 'unread'){
      vm.markItemsRead()
    }
  },
  toggleItemStarred: function() {
    if (vm.itemSelected != null) {
      vm.toggleItemStarred(vm.itemSelectedDetails)
    }
  },
  focusSearch: function() {
    var input = document.getElementById("searchbar")
    input.value = ''
    var ev = new Event('input', {bubbles: true})
    input.dispatchEvent(ev)
    input.focus()
    input.select()
  },
  nextItem(){
    vm.navigateToItem(+1)
  },
  previousItem() {
    vm.navigateToItem(-1)
  },
  nextFeed(){
    vm.navigateToFeed(+1)
  },
  previousFeed() {
    vm.navigateToFeed(-1)
  },
  scrollForward: function() {
    helperFunctions.scrollContent(+1)
  },
  scrollBackward: function() {
    helperFunctions.scrollContent(-1)
  },
  closeItem: function () {
    vm.itemSelected = null
  },
  refreshFeeds: function() {
    vm.fetchAllFeeds()
  },
  showAll() {
    vm.filterSelected = ''
  },
  showUnread() {
    vm.filterSelected = 'unread'
  },
  showStarred() {
    vm.filterSelected = 'starred'
  },
  showAllFeeds() {
    vm.feedSelected = ''
    vm.$nextTick(function() {
      var scroll = document.getElementById('feed-list-scroll')
      var handle = scroll.querySelector('input[type=radio]:checked')
      var target = handle && handle.parentElement
      if (target && scroll) scrollto(target, scroll)
    })
  },
  toggleAllFolders() {
    var allExpanded = vm.folders.every(function(folder) {
      return folder.is_expanded
    })
    var newState = !allExpanded
    vm.folders.forEach(function(folder) {
      folder.is_expanded = newState
      api.folders.update(folder.id, {is_expanded: newState})
    })
  },
  toggleCurrentFolder() {
    var current = vm.current
    if (current.type === 'folder') {
      var folder = current.folder
      folder.is_expanded = !folder.is_expanded
      api.folders.update(folder.id, {is_expanded: folder.is_expanded})
    } else if (current.type === 'feed') {
      var feed = current.feed
      if (feed.folder_id) {
        var folder = vm.foldersById[feed.folder_id]
        if (folder) {
          folder.is_expanded = !folder.is_expanded
          api.folders.update(folder.id, {is_expanded: folder.is_expanded})
        }
      }
    }
  },
  resizeFeedList(delta) {
    vm.resizeFeedList(vm.feedListWidth + delta)
  },
  resizeItemList(delta) {
    vm.resizeItemList(vm.itemListWidth + delta)
  },
}

// If you edit, make sure you update the help modal
var keybindings = {
  "o": shortcutFunctions.openItemLink,
  "v": shortcutFunctions.openItemLink,
  "i": shortcutFunctions.toggleReadability,
  "r": shortcutFunctions.refreshFeeds,
  "R": shortcutFunctions.markAllRead,
  "?": shortcutFunctions.openShortcuts,
  "Escape": shortcutFunctions.closeModal,
  "s": shortcutFunctions.toggleItemStarred,
  "/": shortcutFunctions.focusSearch,
  "j": shortcutFunctions.nextItem,
  "k": shortcutFunctions.previousItem,
  "l": shortcutFunctions.nextFeed,
  "h": shortcutFunctions.previousFeed,
  "f": shortcutFunctions.scrollForward,
  "F": shortcutFunctions.scrollBackward,
  "e": shortcutFunctions.scrollBackward,
  // "e": shortcutFunctions.showAll,
  // "q": shortcutFunctions.closeItem,
  "u": shortcutFunctions.toggleItemRead,
  "1": shortcutFunctions.showUnread,
  "2": shortcutFunctions.showStarred,
  "3": shortcutFunctions.showAll,
  "q": shortcutFunctions.showUnread,
  // "w": shortcutFunctions.showStarred,
  "A": shortcutFunctions.showAllFeeds,
  "a": shortcutFunctions.markAllRead,
  "x": shortcutFunctions.toggleCurrentFolder,
  "c": shortcutFunctions.toggleCurrentFolder,
  "C": shortcutFunctions.toggleAllFolders,
  "X": shortcutFunctions.toggleAllFolders,
  "[": function() { shortcutFunctions.resizeFeedList(-50) },
  "]": function() { shortcutFunctions.resizeFeedList(50) },
  "{": function() { shortcutFunctions.resizeItemList(-50) },
  "}": function() { shortcutFunctions.resizeItemList(50) },
}

var altBindings = {
  'j': shortcutFunctions.nextFeed,
  'k': shortcutFunctions.previousFeed,
  '[': function() { shortcutFunctions.resizeItemList(-30) },
  ']': function() { shortcutFunctions.resizeItemList(30) },
}

var codebindings = {
  "KeyO": shortcutFunctions.openItemLink,
  "KeyV": shortcutFunctions.openItemLink,
  "KeyI": shortcutFunctions.toggleReadability,
  //"r": shortcutFunctions.toggleItemRead,
  //"KeyR": shortcutFunctions.markAllRead,
  "KeyS": shortcutFunctions.toggleItemStarred,
  "Slash": shortcutFunctions.focusSearch,
  "KeyJ": shortcutFunctions.nextItem,
  "KeyK": shortcutFunctions.previousItem,
  "KeyL": shortcutFunctions.nextFeed,
  "KeyH": shortcutFunctions.previousFeed,
  "KeyF": shortcutFunctions.scrollForward,
  "KeyB": shortcutFunctions.scrollBackward,
  // "KeyQ": shortcutFunctions.closeItem,
  "KeyG": shortcutFunctions.toggleItemRead,
  "KeyR": shortcutFunctions.refreshFeeds,
  "KeySlash": shortcutFunctions.openShortcuts,
  "Escape": shortcutFunctions.closeModal,
  "Digit1": shortcutFunctions.showUnread,
  "Digit2": shortcutFunctions.showStarred,
  "Digit3": shortcutFunctions.showAll,
  "KeyQ": shortcutFunctions.showUnread,
  // "KeyW": shortcutFunctions.showStarred,
  // "KeyE": shortcutFunctions.showAll,
  // "KeyA": shortcutFunctions.showAllFeeds,
  "KeyC": shortcutFunctions.toggleCurrentFolder,
}

function isTextBox(element) {
  var tagName = element.tagName.toLowerCase()
  // Input elements that aren't text
  var inputBlocklist = ['button','checkbox','color','file','hidden','image','radio','range','reset','search','submit']

  return tagName === 'textarea' ||
    ( tagName === 'input'
      && inputBlocklist.indexOf(element.getAttribute('type').toLowerCase()) == -1
    )
}

document.addEventListener('keydown',function(event) {
  // in input-like element, Esc only blurs (does not clear)
  if (event.key === 'Escape' && isTextBox(event.target)) {
    event.target.blur()
    return
  }

  // Handle Alt+key shortcuts
  if (event.altKey && !event.ctrlKey && !event.metaKey) {
    var altFn = altBindings[event.key]
    if (altFn) {
      event.preventDefault()
      altFn()
      return
    }
  }

  // Ignore while focused on text or
  // when using modifier keys (to not clash with browser behaviour)
  if (isTextBox(event.target) || event.metaKey || event.ctrlKey || event.altKey) {
    return
  }
  // Handle Space / Shift+Space for scrolling
  if (event.key === ' ') {
    event.preventDefault()
    if (event.shiftKey) {
      shortcutFunctions.scrollBackward()
    } else {
      shortcutFunctions.scrollForward()
    }
    return
  }

  // var keybindFunction = keybindings[event.key] || codebindings[event.code]
  var keybindFunction = keybindings[event.key]
  if (keybindFunction) {
    event.preventDefault()
    keybindFunction()
  }
})
