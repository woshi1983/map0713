// TODO: need to support get browser lang
import { CookieOptions } from 'ngx-cookie';

export const supportedLangs = ['en-us', 'zh-cn', 'zh-tw'];

export const DeFaultLang = 'en-us';
export const languageNames = {
  'en-us': 'English',
  'zh-cn': '中文简体',
  'zh-tw': '中文繁體'
};

/**
 * The default localstorage key used to store current used language preference.
 */
export const LANG_LOCALSTORAGE_KEY = 'map-language';

/*
* the default cookie key used to store current org
 */
export const CURRENT_ORG_KEY = '__map_org';

/*
* the default cookie key used to store current user's role
 */
export const USER_ROLE_KEY = '__user_role';

/*
* used for all login and user info
* */
export const TOKEN_COOKIE_KEY = '_ssid';

export const SECRET_COOKIE_KEY = '_sid';

export const LS_COOKIE_KEY = '_lsid';

export const USER_INFO_COOKIE_KEY = '__user_info';

/*
* used for storage application selected
* */
export const SELECTED_APPLICATION = '_appmgmt_app';

export const enum ALERTTYPE {
  DANGER,
  WARNING,
  INFO,
  SUCCESS
}

/*
*  Feature Gates
* */
export const FEATUREGATES = {
  appMgmt: 'appMgmt',
  appBuild: 'appBuild',
  localUser: 'localUser'
}

export const enum DELETEOBJECTTYPE {
  EMPTY,
  BINARYIMAGE,
  BINARYIMAGEDETAILS,
  BINARYBUILD,
  DOCKERFILEIMAGE,
  DOCKERFILEIMAGEDETAILS,
  DOCKERFILEBUILD,
  SOURCEIMAGE,
  SOURCEIMAGEDETAILS,
  SOURCEBUILD,
  REPO,
  LB,
  CLUSTERPROVIDER
}

/*
* CONFIRM_CANCEL: need confirm with input
* YES_NO: don't need input
* */
export const enum DELETEOBJECTBUTTON {
  CONFIRM_CANCEL, YES_NO, CLOSE, ENABLE_CANCEL, DISABLE_CANCEL
}

export const enum CONFIRMATIONSTATE {
  NA, CONFIRMED, CANCEL
}

// expire one day
export const cookieOption: CookieOptions = {
      path: '/',
      expires: new Date(Date.now() + 3600 * 24 * 1000),
      sameSite: 'strict'
    }


