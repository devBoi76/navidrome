import * as React from 'react'
import {
  Children,
  cloneElement,
  isValidElement,
  useEffect,
  useState,
} from 'react'
import PropTypes from 'prop-types'
import { useTranslate, useGetIdentity } from 'react-admin'
import {
  Tooltip,
  IconButton,
  Popover,
  MenuList,
  Avatar,
  Card,
  CardContent,
  Divider,
  Typography,
} from '@material-ui/core'
import { makeStyles } from '@material-ui/core/styles'
import AccountCircle from '@material-ui/icons/AccountCircle'
import config from '../config'
import authProvider from '../authProvider'
import { startEventStream } from '../eventStream'
import { useDispatch } from 'react-redux'
import { useImageUrl } from '../common'

import KrzyzWKoronie from '../../public/krzyz-w-koronie-256x256.png'

const useStyles = makeStyles((theme) => ({
  user: {},
  button: {
    color: 'inherit',
  },
  avatar: {
    width: theme.spacing(4),
    height: theme.spacing(4),
  },
  username: {
    maxWidth: '11em',
    marginTop: '-0.7em',
    marginBottom: '-1em',
  },
  usernameWrap: {
    whiteSpace: 'nowrap',
    overflow: 'hidden',
    textOverflow: 'ellipsis',
  },
  link: {
    textDecoration: 'none',
    color: theme.palette.primary.main,
    marginLeft: '0.25rem',
  },
  logoImage: {
    height: '1.7em',
    verticalAlign: 'middle',
    marginRight: '0.4rem',
    marginTop: '-0.1rem',
  },
}))

const UserMenu = (props) => {
  const [anchorEl, setAnchorEl] = useState(null)
  const translate = useTranslate()
  const { loaded, identity } = useGetIdentity()
  const classes = useStyles(props)
  const dispatch = useDispatch()

  const { children, label, icon, logout } = props

  useEffect(() => {
    if (config.devActivityPanel) {
      authProvider
        .checkAuth()
        .then(() => startEventStream(dispatch))
        .catch(() => {})
    }
  }, [dispatch])

  if (!logout && !children) return null
  const open = Boolean(anchorEl)

  const handleMenu = (event) => setAnchorEl(event.currentTarget)
  const handleClose = () => setAnchorEl(null)

  return (
    <div className={classes.user}>
      <a
        href="https://www.nastrazy.org"
        rel="noopener noreferrer"
        className={classes.link}
      >
        <img className={classes.logoImage} src={KrzyzWKoronie} />
        Nastrazy.org
      </a>
    </div>
  )
}

UserMenu.propTypes = {
  children: PropTypes.node,
  label: PropTypes.string.isRequired,
  logout: PropTypes.element,
}

UserMenu.defaultProps = {
  label: 'menu.settings',
  icon: <AccountCircle />,
}

export default UserMenu
