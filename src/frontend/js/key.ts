import App from "./pages/App.vue";

export function setupKeybindings(vm: InstanceType<typeof App>) {
  var helperFunctions = {
    scrollContent(direction: number) {
      var padding = 40;
      var scroll = document.querySelector(".content");
      if (!scroll) return;

      var height = scroll.getBoundingClientRect().height;
      var newpos = scroll.scrollTop + (height - padding) * direction;

      if (typeof scroll.scrollTo == "function") {
        scroll.scrollTo({ top: newpos, left: 0, behavior: "smooth" });
      } else {
        scroll.scrollTop = newpos;
      }
    },
  };
  var shortcutFunctions = {
    openItemLink() {
      if (vm.itemSelectedDetails && vm.itemSelectedDetails.link) {
        window.open(vm.itemSelectedDetails.link, "_blank", "noopener,noreferrer");
      }
    },
    scrollBackwardAlias() {
      helperFunctions.scrollContent(-1);
    },
    toggleReadability() {
      vm.toggleReadability();
    },
    refreshFeeds() {
      vm.fetchAllFeeds();
    },
    toggleItemRead() {
      if (vm.itemSelectedDetails !== null) {
        vm.toggleItemRead(vm.itemSelectedDetails);
      }
    },
    markAllRead() {
      // same condition as 'Mark all read button'
      if (vm.filterSelected == "unread") {
        vm.markItemsRead();
      }
    },
    toggleItemStarred() {
      if (vm.itemSelectedDetails != null) {
        vm.toggleItemStarred(vm.itemSelectedDetails);
      }
    },
    focusSearch() {
      document.getElementById("searchbar")?.focus();
    },
    nextItem() {
      vm.navigateToItem(+1);
    },
    previousItem() {
      vm.navigateToItem(-1);
    },
    nextFeed() {
      vm.navigateToFeed(+1);
    },
    previousFeed() {
      vm.navigateToFeed(-1);
    },
    scrollForward() {
      helperFunctions.scrollContent(+1);
    },
    scrollBackward() {
      helperFunctions.scrollContent(-1);
    },
    closeItem() {
      vm.itemSelected = null;
    },
    showShortcuts() {
      vm.showModal = "shortcuts";
    },
    showAll() {
      vm.filterSelected = "";
    },
    showUnread() {
      vm.filterSelected = "unread";
    },
    showStarred() {
      vm.filterSelected = "starred";
    },
    toggleFeedList() {
      vm.toggleFeedList();
    },
    markAllItemsRead() {
      vm.markAllItemsRead();
    },
    showAllFeeds() {
      vm.showAllFeeds();
    },
    toggleCurrentFolder() {
      vm.toggleCurrentFolder();
    },
    collapseAllFolders() {
      vm.toggleAllFolders(false);
    },
    expandAllFolders() {
      vm.toggleAllFolders(true);
    },
    resizeFeedList(delta: number) {
      vm.resizeFeedList(vm.feedListWidth + delta);
    },
    resizeItemList(delta: number) {
      vm.resizeItemList(vm.itemListWidth + delta);
    },
  };

  // If you edit, make sure you update the help modal
  var keybindings = {
    o: shortcutFunctions.openItemLink,
    v: shortcutFunctions.openItemLink,
    i: shortcutFunctions.toggleReadability,
    r: shortcutFunctions.refreshFeeds,
    R: shortcutFunctions.markAllRead,
    s: shortcutFunctions.toggleItemStarred,
    "/": shortcutFunctions.focusSearch,
    j: shortcutFunctions.nextItem,
    k: shortcutFunctions.previousItem,
    l: shortcutFunctions.nextFeed,
    h: shortcutFunctions.previousFeed,
    f: shortcutFunctions.scrollForward,
    b: shortcutFunctions.scrollBackward,
    e: shortcutFunctions.scrollBackwardAlias,
    q: shortcutFunctions.closeItem,
    "?": shortcutFunctions.showShortcuts,
    "1": shortcutFunctions.showUnread,
    "2": shortcutFunctions.showStarred,
    "3": shortcutFunctions.showAll,
    u: shortcutFunctions.toggleItemRead,
    w: shortcutFunctions.toggleFeedList,
    a: shortcutFunctions.markAllItemsRead,
    A: shortcutFunctions.showAllFeeds,
    x: shortcutFunctions.toggleCurrentFolder,
    c: shortcutFunctions.toggleCurrentFolder,
    C: shortcutFunctions.collapseAllFolders,
    X: shortcutFunctions.expandAllFolders,
    "[": function () {
      shortcutFunctions.resizeFeedList(-50);
    },
    "]": function () {
      shortcutFunctions.resizeFeedList(+50);
    },
    "{": function () {
      shortcutFunctions.resizeItemList(-50);
    },
    "}": function () {
      shortcutFunctions.resizeItemList(+50);
    },
  } as Record<string, CallableFunction>;

  var codebindings = {
    KeyO: shortcutFunctions.openItemLink,
    KeyV: shortcutFunctions.openItemLink,
    KeyI: shortcutFunctions.toggleReadability,
    KeyR: shortcutFunctions.refreshFeeds,
    //"KeyR": shortcutFunctions.markAllRead,
    KeyS: shortcutFunctions.toggleItemStarred,
    Slash: shortcutFunctions.focusSearch,
    KeyJ: shortcutFunctions.nextItem,
    KeyK: shortcutFunctions.previousItem,
    KeyL: shortcutFunctions.nextFeed,
    KeyH: shortcutFunctions.previousFeed,
    KeyF: shortcutFunctions.scrollForward,
    KeyB: shortcutFunctions.scrollBackward,
    KeyE: shortcutFunctions.scrollBackwardAlias,
    KeyQ: shortcutFunctions.closeItem,
    Digit1: shortcutFunctions.showUnread,
    Digit2: shortcutFunctions.showStarred,
    Digit3: shortcutFunctions.showAll,
    KeyU: shortcutFunctions.toggleItemRead,
    KeyW: shortcutFunctions.toggleFeedList,
    KeyA: shortcutFunctions.markAllItemsRead,
    KeyX: shortcutFunctions.toggleCurrentFolder,
    KeyC: shortcutFunctions.toggleCurrentFolder,
  } as Record<string, CallableFunction>;

  function isTextBox(element: Element) {
    var tagName = element.tagName.toLowerCase();
    // Input elements that aren't text
    var inputBlocklist = [
      "button",
      "checkbox",
      "color",
      "file",
      "hidden",
      "image",
      "radio",
      "range",
      "reset",
      "search",
      "submit",
    ];

    return (
      tagName === "textarea" ||
      (tagName === "input" &&
        inputBlocklist.indexOf(element.getAttribute("type")?.toLowerCase() || "") == -1)
    );
  }

  document.addEventListener("keydown", function (event) {
    // Ignore while focused on text or
    // when using modifier keys (to not clash with browser behaviour)
    if (isTextBox(event.target as Element) || event.metaKey || event.ctrlKey || event.altKey) {
      return;
    }
    var keybindFunction = keybindings[event.key] || codebindings[event.code];
    if (keybindFunction) {
      event.preventDefault();
      keybindFunction();
    }
  });
}
