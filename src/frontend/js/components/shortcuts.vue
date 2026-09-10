<template>
  <div>
    <p class="cursor-default">
      <b>{{ $t("keyboard_shortcuts") }}</b>
    </p>
    <div class="d-flex flex-column mt-2">
      <div class="d-flex py-1">
        <div class="w-25"><kbd>1</kbd> <kbd>2</kbd> <kbd>3</kbd></div>
        <div class="w-75">{{ $t("kb_show_filters") }}</div>
      </div>
      <div class="d-flex py-1">
        <div class="w-25"><kbd>/</kbd></div>
        <div class="w-75">{{ $t("kb_focus_search") }}</div>
      </div>

      <div class="py-3"></div>

      <div class="d-flex py-1">
        <div class="w-25"><kbd>j</kbd> <kbd>k</kbd></div>
        <div class="w-75">{{ $t("kb_next_prev_article") }}</div>
      </div>
      <div class="d-flex py-1">
        <div class="w-25"><kbd>l</kbd> <kbd>h</kbd></div>
        <div class="w-75">{{ $t("kb_next_prev_feed") }}</div>
      </div>
      <div class="d-flex py-1">
        <div class="w-25"><kbd>q</kbd></div>
        <div class="w-75">{{ $t("kb_close_article") }}</div>
      </div>

      <div class="py-3"></div>

      <div class="d-flex py-1">
        <div class="w-25"><kbd>R</kbd></div>
        <div class="w-75">{{ $t("kb_mark_all_read") }}</div>
      </div>
      <div class="d-flex py-1">
        <div class="w-25"><kbd>r</kbd></div>
        <div class="w-75">{{ $t("kb_mark_read") }}</div>
      </div>
      <div class="d-flex py-1">
        <div class="w-25"><kbd>s</kbd></div>
        <div class="w-75">{{ $t("kb_mark_starred") }}</div>
      </div>
      <div class="d-flex py-1">
        <div class="w-25"><kbd>o</kbd></div>
        <div class="w-75">{{ $t("kb_open_link") }}</div>
      </div>
      <div class="d-flex py-1">
        <div class="w-25"><kbd>i</kbd></div>
        <div class="w-75">{{ $t("kb_read_here") }}</div>
      </div>
      <div class="d-flex py-1">
        <div class="w-25"><kbd>f</kbd> <kbd>b</kbd></div>
        <div class="w-75">{{ $t("kb_scroll_content") }}</div>
      </div>

      <div class="py-3"></div>

      <div class="d-flex py-1">
        <div class="w-25"><kbd>w</kbd></div>
        <div class="w-75">{{ $t("kb_toggle_feed_list") }}</div>
      </div>
      <div class="d-flex py-1">
        <div class="w-25"><kbd>a</kbd></div>
        <div class="w-75">{{ $t("kb_mark_all_read_everywhere") }}</div>
      </div>
      <div class="d-flex py-1">
        <div class="w-25"><kbd>A</kbd></div>
        <div class="w-75">{{ $t("kb_show_all_feeds") }}</div>
      </div>
      <div class="d-flex py-1">
        <div class="w-25"><kbd>x</kbd> <kbd>c</kbd></div>
        <div class="w-75">{{ $t("kb_toggle_current_folder") }}</div>
      </div>
      <div class="d-flex py-1">
        <div class="w-25"><kbd>C</kbd> <kbd>X</kbd></div>
        <div class="w-75">{{ $t("kb_toggle_all_folders") }}</div>
      </div>
      <div class="d-flex py-1">
        <div class="w-25"><kbd>[</kbd> <kbd>]</kbd> <kbd>{</kbd> <kbd>}</kbd></div>
        <div class="w-75">{{ $t("kb_resize_columns") }}</div>
      </div>
    </div>
  </div>
</template>
